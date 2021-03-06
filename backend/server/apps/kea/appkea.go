package kea

import (
	"context"
	"time"

	errors "github.com/pkg/errors"
	log "github.com/sirupsen/logrus"

	"isc.org/stork/server/agentcomm"
	dbops "isc.org/stork/server/database"
	dbmodel "isc.org/stork/server/database/model"
	storkutil "isc.org/stork/util"
)

const (
	dhcp4 = "dhcp4"
	dhcp6 = "dhcp6"
	d2    = "d2"
)

// Get list of hooks for all DHCP daemons of the given Kea application.
func GetDaemonHooks(dbApp *dbmodel.App) map[string][]string {
	hooksByDaemon := make(map[string][]string)

	// go through response list with configs from each daemon and retrieve their hooks lists
	for _, dmn := range dbApp.Daemons {
		if dmn.KeaDaemon == nil || dmn.KeaDaemon.Config == nil {
			continue
		}

		libraries := dmn.KeaDaemon.Config.GetHooksLibraries()
		hooks := []string{}
		for _, library := range libraries {
			hooks = append(hooks, library.Library)
		}
		hooksByDaemon[dmn.Name] = hooks
	}

	return hooksByDaemon
}

// === CA config-get response structs ================================================

type SocketData struct {
	SocketName string `json:"socket-name"`
	SocketType string `json:"socket-type"`
}

type ControlSocketsData struct {
	D2      *SocketData
	Dhcp4   *SocketData
	Dhcp6   *SocketData
	NetConf *SocketData
}

type ControlAgentData struct {
	ControlSockets *ControlSocketsData `json:"control-sockets"`
}

type CAConfigGetRespArgs struct {
	ControlAgent *ControlAgentData `json:"Control-agent"`
}

type CAConfigGetResponse struct {
	agentcomm.KeaResponseHeader
	Arguments *CAConfigGetRespArgs
}

// === version-get response structs ===============================================

type VersionGetRespArgs struct {
	Extended string
}

type VersionGetResponse struct {
	agentcomm.KeaResponseHeader
	Arguments *VersionGetRespArgs `json:"arguments,omitempty"`
}

// Get state of Kea application Control Agent using ForwardToKeaOverHTTP function.
// The state, that is stored into dbApp, includes: version and config of CA.
// It also returns:
// - list of all Kea daemons
// - list of DHCP daemons (dhcpv4 and/or dhcpv6)
func getStateFromCA(ctx context.Context, agents agentcomm.ConnectedAgents, caURL string, dbApp *dbmodel.App, daemonsMap map[string]*dbmodel.Daemon) (agentcomm.KeaDaemons, agentcomm.KeaDaemons, error) {
	// prepare the command to get config and version from CA
	cmds := []*agentcomm.KeaCommand{
		{
			Command: "version-get",
		},
		{
			Command: "config-get",
		},
	}

	// get version and config from CA
	versionGetResp := []VersionGetResponse{}
	caConfigGetResp := []CAConfigGetResponse{}

	cmdsResult, err := agents.ForwardToKeaOverHTTP(ctx, dbApp.Machine.Address, dbApp.Machine.AgentPort, caURL, cmds, &versionGetResp, &caConfigGetResp)
	if err != nil {
		return nil, nil, err
	}
	if cmdsResult.Error != nil {
		return nil, nil, cmdsResult.Error
	}

	// process the response from CA
	daemonsMap["ca"] = &dbmodel.Daemon{
		Name:   "ca",
		Active: true,
	}

	if cmdsResult.CmdsErrors[0] == nil {
		vRsp := versionGetResp[0]
		dmn := daemonsMap["ca"]
		if vRsp.Result != 0 {
			dmn.Active = false
			log.Warnf("problem with version-get from CA: %s", vRsp.Text)
		} else {
			dmn.Version = vRsp.Text
			dbApp.Meta.Version = vRsp.Text
			if vRsp.Arguments != nil {
				dmn.ExtendedVersion = vRsp.Arguments.Extended
			}
		}
	} else {
		log.Warnf("problem with version-get response from CA: %s", cmdsResult.CmdsErrors[0])
	}

	allDaemons := make(agentcomm.KeaDaemons)
	dhcpDaemons := make(agentcomm.KeaDaemons)
	if caConfigGetResp[0].Arguments.ControlAgent.ControlSockets != nil {
		if caConfigGetResp[0].Arguments.ControlAgent.ControlSockets.Dhcp4 != nil {
			allDaemons[dhcp4] = true
			dhcpDaemons[dhcp4] = true
		}
		if caConfigGetResp[0].Arguments.ControlAgent.ControlSockets.Dhcp6 != nil {
			allDaemons[dhcp6] = true
			dhcpDaemons[dhcp6] = true
		}
		if caConfigGetResp[0].Arguments.ControlAgent.ControlSockets.D2 != nil {
			allDaemons[d2] = true
		}
	}

	return allDaemons, dhcpDaemons, nil
}

// Get state of Kea application daemons (beside Control Agent) using ForwardToKeaOverHTTP function.
// The state, that is stored into dbApp, includes: version, config and runtime state of indicated Kea daemons.
func getStateFromDaemons(ctx context.Context, agents agentcomm.ConnectedAgents, caURL string, dbApp *dbmodel.App, daemonsMap map[string]*dbmodel.Daemon, allDaemons agentcomm.KeaDaemons, dhcpDaemons agentcomm.KeaDaemons) error {
	now := storkutil.UTCNow()

	// issue 3 commands to Kea daemons at once to get their state
	cmds := []*agentcomm.KeaCommand{
		{
			Command: "version-get",
			Daemons: &allDaemons,
		},
		{
			Command: "status-get",
			Daemons: &dhcpDaemons,
		},
		{
			Command: "config-get",
			Daemons: &allDaemons,
		},
	}

	versionGetResp := []VersionGetResponse{}
	statusGetResp := []StatusGetResponse{}
	configGetResp := []agentcomm.KeaResponse{}

	cmdsResult, err := agents.ForwardToKeaOverHTTP(ctx, dbApp.Machine.Address, dbApp.Machine.AgentPort, caURL, cmds, &versionGetResp, &statusGetResp, &configGetResp)
	if err != nil {
		return err
	}
	if cmdsResult.Error != nil {
		return cmdsResult.Error
	}

	for name := range allDaemons {
		daemonsMap[name] = dbmodel.NewKeaDaemon(name, true)
	}

	// process version-get responses
	err = cmdsResult.CmdsErrors[0]
	if err != nil {
		log.Warnf("problem with version-get response: %s", err)
	} else {
		for _, vRsp := range versionGetResp {
			dmn := daemonsMap[vRsp.Daemon]
			if vRsp.Result != 0 {
				dmn.Active = false
				log.Warnf("problem with version-get and kea daemon %s: %s", vRsp.Daemon, vRsp.Text)
				continue
			}

			dmn.Version = vRsp.Text
			if vRsp.Arguments != nil {
				dmn.ExtendedVersion = vRsp.Arguments.Extended
			}
		}
	}

	// process status-get responses
	err = cmdsResult.CmdsErrors[1]
	if err != nil {
		log.Warnf("problem with status-get response: %s", err)
	} else {
		for _, sRsp := range statusGetResp {
			dmn := daemonsMap[sRsp.Daemon]
			if sRsp.Result != 0 {
				dmn.Active = false
				log.Warnf("problem with status-get and kea daemon %s: %s", sRsp.Daemon, sRsp.Text)
				continue
			}

			if sRsp.Arguments != nil {
				dmn.Uptime = sRsp.Arguments.Uptime
				dmn.ReloadedAt = now.Add(time.Second * time.Duration(-sRsp.Arguments.Reload))
				// TODO: HA status
			}
		}
	}

	// process config-get responses
	err = cmdsResult.CmdsErrors[2]
	if err != nil {
		log.Warnf("problem with config-get response: %s", err)
	} else {
		for _, cRsp := range configGetResp {
			dmn := daemonsMap[cRsp.Daemon]
			if cRsp.Result != 0 {
				dmn.Active = false
				log.Warnf("problem with config-get and kea daemon %s: %s", cRsp.Daemon, cRsp.Text)
				continue
			}

			dmn.KeaDaemon.Config = dbmodel.NewKeaConfig(cRsp.Arguments)
		}
	}

	return nil
}

// Get state of Kea application daemons using ForwardToKeaOverHTTP function.
// The state that is stored into dbApp includes: version, config and runtime state of indicated Kea daemons.
func GetAppState(ctx context.Context, agents agentcomm.ConnectedAgents, dbApp *dbmodel.App) {
	// prepare URL to CA
	ctrlPoint, err := dbApp.GetAccessPoint(dbmodel.AccessPointControl)
	if err != nil {
		log.Warnf("problem with getting kea access control point: %s", err)
		return
	}
	caURL := storkutil.HostWithPortURL(ctrlPoint.Address, ctrlPoint.Port)

	ctx2, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	// get state from CA
	daemonsMap := map[string]*dbmodel.Daemon{}
	allDaemons, dhcpDaemons, err := getStateFromCA(ctx2, agents, caURL, dbApp, daemonsMap)

	// if not problems then now get state from the rest of Kea daemons
	if err == nil {
		err = getStateFromDaemons(ctx2, agents, caURL, dbApp, daemonsMap, allDaemons, dhcpDaemons)
		if err != nil {
			log.Warnf("problem with getting state from kea daemons: %s", err)
		}
	} else {
		log.Warnf("problem with getting state from kea CA: %s", err)
	}

	// store all collected details in app db record
	dbApp.Active = true
	for name := range daemonsMap {
		dmn := daemonsMap[name]
		// if all daemons are active then whole app is active
		dbApp.Active = dbApp.Active && dmn.Active

		dbApp.Daemons = append(dbApp.Daemons, dmn)
	}
}

// Inserts or updates information about Kea app in the database. Next, it extracts
// Kea's configurations and uses to either update or create new shared networks,
// subnets and pools. Finally, the relations between the subnets and the Kea app
// are created. Note that multiple apps can be associated with the same subnet.
func CommitAppIntoDB(db *dbops.PgDB, app *dbmodel.App) error {
	// Go over the shared networks and subnets stored in the Kea configuration
	// and match them with the existing entires in the database. If some of
	// the shared networks or subnets do not exist they are instantiated and
	// returned here.
	networks, subnets, err := DetectNetworks(db, app)
	if err != nil {
		err = errors.Wrapf(err, "unable to detect subnets and shared networks for Kea app with id %d", app.ID)
		return err
	}

	// Go over the global reservations stored in the Kea configuration and
	// match them with the existing global hosts.
	globalHosts, err := detectGlobalHostsFromConfig(db, app)
	if err != nil {
		err = errors.Wrapf(err, "unable to detect global host reservations for Kea app with id %d", app.ID)
		return err
	}

	// Begin transaction.
	tx, rollback, commit, err := dbops.Transaction(db)
	if err != nil {
		return err
	}
	defer rollback()

	if app.ID == 0 {
		// New app, insert it.
		err = dbmodel.AddApp(tx, app)
	} else {
		// Existing app, update it if needed.
		err = dbmodel.UpdateApp(tx, app)
	}

	if err != nil {
		return err
	}

	// For the given app, iterate over the networks and subnets and update their
	// global instances accordingly in the database.
	err = dbmodel.CommitNetworksIntoDB(tx, networks, subnets, app, 1)
	if err != nil {
		return err
	}

	// For the given app, iterate over the global hosts and update their instances
	// in the database or insert them into the database.
	err = dbmodel.CommitGlobalHostsIntoDB(tx, globalHosts, app, "config", 1)
	if err != nil {
		return err
	}

	for _, daemon := range app.Daemons {
		// Check what HA services the daemon belongs to.
		services := DetectHAServices(db, daemon)

		// For the given daemon, iterate over the services and add/update them in the
		// database.
		err = dbmodel.CommitServicesIntoDB(tx, services, daemon)
		if err != nil {
			return err
		}
	}

	// Commit the changes if everything went fine.
	err = commit()
	return err
}
