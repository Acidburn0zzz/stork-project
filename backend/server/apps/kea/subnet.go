package kea

import (
	log "github.com/sirupsen/logrus"

	dbops "isc.org/stork/server/database"
	dbmodel "isc.org/stork/server/database/model"
)

// Checks whether the given shared network exists already. It iterates over the
// slice of existing networks. If the network seems to be matching one of them,
// the shared network instance along with all subnets is fetched from the
// database and returned to the caller.
func sharedNetworkExists(db *dbops.PgDB, network *dbmodel.SharedNetwork, existingNetworks []dbmodel.SharedNetwork) (*dbmodel.SharedNetwork, error) {
	for _, existing := range existingNetworks {
		// todo: this is logic should be extended to perform some more sophisticated
		// matching of a shared network with existing shared networks. For now,
		// we only match by the shared network name and we do not resolve any
		// conflicts. This should change soon.
		if existing.Name == network.Name {
			// Get the subnets included in this shared network.
			dbNetwork, err := dbmodel.GetSharedNetworkWithSubnets(db, existing.ID)
			if err != nil {
				return nil, err
			}
			return dbNetwork, nil
		}
	}
	return nil, nil
}

// Checks whether the given subnet exists already. It iterates over the slice
// of existing subnets. If the subnet matches one of them the index of the
// subnet is returned to the caller.
func subnetExists(subnet *dbmodel.Subnet, existingSubnets []dbmodel.Subnet) (bool, int) {
	// todo: this logic should be extended to perform some more sophisticated
	// matching of the subnet with existing subnets. For now, we only match by
	// the subnet prefix and we do not resolve any conflicts. This should
	// change soon.
	for i, existing := range existingSubnets {
		if existing.Prefix == subnet.Prefix {
			return true, i
		}
	}
	return false, 0
}

// For a given Kea configuration it detects the shared networks matching this
// configuration. All existing shared network matching the given configuration
// are returned as they are. If there is no match a new shared network instance
// is returned.
func detectSharedNetworks(db *dbops.PgDB, config *dbmodel.KeaConfig, family int, app *dbmodel.App) (networks []dbmodel.SharedNetwork, err error) {
	// Get all shared networks and the subnets within those networks from the
	// application configuration.
	if networkList, ok := config.GetTopLevelList("shared-networks"); ok {
		// If there are no shared networks there is nothing to do.
		if len(networkList) == 0 {
			return networks, nil
		}

		// We have to match the configured shared networks with the ones we
		// already have in the database.
		dbNetworks, err := dbmodel.GetAllSharedNetworks(db, family)
		if err != nil {
			return []dbmodel.SharedNetwork{}, err
		}

		// For each network in the app's configuration we will do such matching.
		for _, n := range networkList {
			if networkMap, ok := n.(map[string]interface{}); ok {
				// Parse the configured network.
				network, err := dbmodel.NewSharedNetworkFromKea(&networkMap, family)
				if err != nil {
					log.Warnf("skipping invalid shared network: %v", err)
					continue
				}
				dbNetwork, err := sharedNetworkExists(db, network, dbNetworks)
				if err != nil {
					return []dbmodel.SharedNetwork{}, err
				}
				if dbNetwork != nil {
					// Go over the configured subnets and see if they belong to that
					// shared network already.
					for _, s := range network.Subnets {
						subnet := s
						ok, idx := subnetExists(&subnet, dbNetwork.Subnets)
						if !ok {
							dbNetwork.Subnets = append(dbNetwork.Subnets, subnet)
						} else {
							// Subnet already exists and may contain some hosts. Let's
							// merge the hosts from the new subnet into the existing subnet.
							hosts, err := mergeSubnetHosts(db, &dbNetwork.Subnets[idx], &subnet, app)
							if err != nil {
								log.Warnf("skipping hosts for subnet %s after hosts merge failure: %v",
									subnet.Prefix, err)
								continue
							}
							dbNetwork.Subnets[idx].Hosts = hosts
						}
					}
					networks = append(networks, *dbNetwork)
				} else {
					networks = append(networks, *network)
				}
			}
		}
	}
	return networks, nil
}

// For a given Kea configuration it detects the top-level subnets matching
// this configuration. All existing subnets matching the given configuration
// are returned as they are. If there is no match a new subnet instance is
// returned.
func detectSubnets(db *dbops.PgDB, config *dbmodel.KeaConfig, family int, app *dbmodel.App) (subnets []dbmodel.Subnet, err error) {
	subnetParamName := "subnet4"
	if family == 6 {
		subnetParamName = "subnet6"
	}

	// Get top level subnets not associated with any shared networks.
	if subnetList, ok := config.GetTopLevelList(subnetParamName); ok {
		// Nothing to do if no subnets are configured.
		if len(subnetList) == 0 {
			return subnets, nil
		}

		// Fetch all top-level subnets from the database to perform matching. For now
		// it is better to get all of them because this is just a single query rather
		// than many but in the future we should probably revise that when the number
		// of subnets grows.
		dbSubnets, err := dbmodel.GetAllSubnets(db, family)
		if err != nil {
			return []dbmodel.Subnet{}, err
		}

		// Iterate over the configured subnets.
		for _, s := range subnetList {
			if subnetMap, ok := s.(map[string]interface{}); ok {
				// Parse the configured subnet.
				subnet, err := dbmodel.NewSubnetFromKea(&subnetMap)
				if err != nil {
					log.Warnf("skipping invalid subnet: %v", err)
					continue
				}
				exists, index := subnetExists(subnet, dbSubnets)
				if exists {
					subnets = append(subnets, dbSubnets[index])
					// Subnet already exists and may contain some hosts. Let's
					// merge the hosts from the new subnet into the existing subnet.
					hosts, err := mergeSubnetHosts(db, &dbSubnets[index], subnet, app)
					if err != nil {
						log.Warnf("skipping hosts for subnet %s after hosts merge failure: %v",
							subnet.Prefix, err)
						continue
					}
					// Assign merged hosts to the subnet.
					subnets[len(subnets)-1].Hosts = hosts
				} else {
					subnets = append(subnets, *subnet)
				}
			}
		}
	}
	return subnets, err
}

// For a given Kea application it detects the shared networks and subnets this
// Kea instance has configured. It returns both DHCPv4 and DHCPv6 shared networks
// and subnets. The returned shared networks contain the subnets belonging to
// the shared networks.
func DetectNetworks(db *dbops.PgDB, app *dbmodel.App) (networks []dbmodel.SharedNetwork, subnets []dbmodel.Subnet, err error) {
	// If this is not Kea application there is nothing to do.
	if app.Type != dbmodel.AppTypeKea {
		return networks, subnets, nil
	}

	for _, d := range app.Daemons {
		if d.KeaDaemon == nil || d.KeaDaemon.Config == nil {
			continue
		}

		var family int
		switch d.Name {
		case dhcp4:
			family = 4
		case dhcp6:
			family = 6
		default:
			continue
		}

		// Detect shared networks and the subnets.
		detectedNetworks, err := detectSharedNetworks(db, d.KeaDaemon.Config, family, app)
		if err != nil {
			return networks, subnets, err
		}
		networks = append(networks, detectedNetworks...)

		// Detect top level subnets.
		detectedSubnets, err := detectSubnets(db, d.KeaDaemon.Config, family, app)
		if err != nil {
			return []dbmodel.SharedNetwork{}, subnets, err
		}
		subnets = append(subnets, detectedSubnets...)
	}
	return networks, subnets, nil
}
