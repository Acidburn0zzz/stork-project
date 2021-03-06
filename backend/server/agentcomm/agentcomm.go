package agentcomm

import (
	"context"
	"sync"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"

	agentapi "isc.org/stork/api"
)

// Settings specific to communication with Agents
type AgentsSettings struct {
}

// Runtime information about the agent, e.g. connection.
type Agent struct {
	Address  string
	Client   agentapi.AgentClient
	GrpcConn *grpc.ClientConn
}

// Prepare gRPC connection to agent.
func (agent *Agent) MakeGrpcConnection() error {
	// If there is any old connection then clean it up
	if agent.GrpcConn != nil {
		agent.GrpcConn.Close()
	}

	// Setup new connection
	var opts []grpc.DialOption
	opts = append(opts, grpc.WithInsecure())

	grpcConn, err := grpc.Dial(agent.Address, opts...)
	if err != nil {
		return errors.Wrapf(err, "problem with dial to agent %s", agent.Address)
	}

	agent.Client = agentapi.NewAgentClient(grpcConn)
	agent.GrpcConn = grpcConn

	return nil
}

// Interface for interacting with Agents via gRPC.
type ConnectedAgents interface {
	Shutdown()
	GetConnectedAgent(address string) (*Agent, error)
	GetState(ctx context.Context, address string, agentPort int64) (*State, error)
	ForwardRndcCommand(ctx context.Context, agentAddress string, agentPort int64, rndcSettings Bind9Control, command string) (*RndcOutput, error)
	ForwardToNamedStats(ctx context.Context, agentAddress string, agentPort int64, statsURL string, statsOutput interface{}) error
	ForwardToKeaOverHTTP(ctx context.Context, agentAddress string, agentPort int64, caURL string, commands []*KeaCommand, cmdResponses ...interface{}) (*KeaCmdsResult, error)
}

// Agents management map. It tracks Agents currently connected to the Server.
type connectedAgentsData struct {
	Settings     *AgentsSettings
	AgentsMap    map[string]*Agent
	CommLoopReqs chan *commLoopReq
	DoneCommLoop chan bool
	Wg           *sync.WaitGroup
}

// Create new ConnectedAgents objects.
func NewConnectedAgents(settings *AgentsSettings) ConnectedAgents {
	agents := connectedAgentsData{
		Settings:     settings,
		AgentsMap:    make(map[string]*Agent),
		CommLoopReqs: make(chan *commLoopReq),
		DoneCommLoop: make(chan bool),
		Wg:           &sync.WaitGroup{},
	}

	agents.Wg.Add(1)
	go agents.communicationLoop()

	return &agents
}

// Shutdown agents in agents map.
func (agents *connectedAgentsData) Shutdown() {
	log.Printf("Stopping communication with agents")
	for _, agent := range agents.AgentsMap {
		agent.GrpcConn.Close()
	}

	close(agents.CommLoopReqs)
	agents.DoneCommLoop <- true
	agents.Wg.Wait()
	log.Printf("Stopped communication with agents")
}

// Get Agent object by its address.
func (agents *connectedAgentsData) GetConnectedAgent(address string) (*Agent, error) {
	// Look for agent in Agents map and if found then return it
	agent, ok := agents.AgentsMap[address]
	if ok {
		log.Printf("connecting to existing agent on %v", address)
		return agent, nil
	}

	// Agent not found so allocate agent and prepare connection
	agent = new(Agent)
	agent.Address = address
	err := agent.MakeGrpcConnection()
	if err != nil {
		return nil, err
	}

	// Store it in Agents map
	agents.AgentsMap[address] = agent
	log.Printf("connecting to new agent on %v", address)

	return agent, nil
}
