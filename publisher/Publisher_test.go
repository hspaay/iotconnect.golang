package publisher_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/iotdomain/iotdomain-go/messaging"
	"github.com/iotdomain/iotdomain-go/publisher"
	"github.com/iotdomain/iotdomain-go/types"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const domain = "test"

const node1ID = "node1"
const node1AliasID = "alias1"
const publisher1ID = "publisher1"
const publisher2ID = "publisher2"
const configFolder = "../test"
const cacheFolder = "../test/cache"

var node1Base = fmt.Sprintf("%s/%s/%s", domain, publisher1ID, node1ID)
var node1Addr = node1Base + "/$node"
var node1Alias = fmt.Sprintf("%s/%s/%s", domain, publisher1ID, node1AliasID)

// var node1 = nodes.NewNode(domain, publisher1ID, node1ID, types.NodeTypeUnknown)
var node1ConfigureAddr = node1Base + "/$configure"
var node1InputType = types.InputTypeSwitch
var node1InputAddr = node1Base + "/switch/0/$input"
var node1Output1Addr = node1Base + "/switch/0/$output"

var node1Output1Type = types.OutputTypeSwitch // "switch"
var node1Output1Instance = "0"

var node1AliasOutput1Addr = node1Alias + "/switch/0/$output"
var node1valueAddr = node1Base + "/switch/0/$raw"
var node1latestAddr = node1Base + "/switch/0/$latest"
var node1historyAddr = node1Base + "/switch/0/$history"

// var node1Input1 = inputs.NewInput(domain, publisher1ID, node1ID, "switch", "0")
// var node1Output1 = outputs.NewOutput(domain, publisher1ID, node1ID, "switch", "0")
var pubAddr = fmt.Sprintf("%s/%s/$identity", domain, publisher1ID)
var pub2Addr = fmt.Sprintf("%s/%s/$identity", domain, publisher2ID)

var msgConfig *messaging.MessengerConfig = &messaging.MessengerConfig{Domain: domain}

func TestNewPublisher(t *testing.T) {
	var testMessenger = messaging.NewDummyMessenger(msgConfig)

	pub1 := publisher.NewPublisher(domain, publisher1ID, configFolder, false, testMessenger)
	require.NotNil(t, pub1, "Failed creating publisher")

	assert.NotNil(t, pub1.GetIdentity, "Missing publisher identity")
	assert.NotEmpty(t, pub1.Address(), "Missing publisher address")
}

type s struct{ Item1 string }

func TestNewAppPublisher(t *testing.T) {
	appID := "testapp"
	var appConfig struct{ Item1 string }
	appPub, err := publisher.NewAppPublisher(appID, configFolder, cacheFolder, &appConfig, false)
	assert.NotNil(t, appPub)
	assert.Error(t, err) // no messenger config
}

func TestStartStop(t *testing.T) {
	var testMessenger = messaging.NewDummyMessenger(msgConfig)
	var pollHandlerCalled = false

	pub1 := publisher.NewPublisher(domain, publisher1ID, configFolder, false, testMessenger)
	pub1.SetPollInterval(1, func(pub *publisher.Publisher) {
		pollHandlerCalled = true
	})
	pub1.Start()
	time.Sleep(time.Second)
	pub1.Stop()

	// test runner doesn't like a sigint
	// go syscall.Kill(syscall.Getpid(), syscall.SIGINT)
	// pub1.WaitForSignal()

	// should be no problem to stop again
	pub1.Stop()

	assert.True(t, pollHandlerCalled)
}

func TestSetLogging(t *testing.T) {
	var logFile = "/tmp/iotdomain-go.log"
	var testMessenger = messaging.NewDummyMessenger(msgConfig)
	pub1 := publisher.NewPublisher(domain, publisher1ID, configFolder, false, testMessenger)
	pub1.SetLogging("debug", "")
	logrus.Debug("Hello from logger")

	pub1.SetLogging("info", logFile)
	logrus.Info("Hello from logger")

}

func TestLoadNodes(t *testing.T) {
	const device1ID = "device1"
	const device1Type = types.NodeTypeAVReceiver

	var testMessenger = messaging.NewDummyMessenger(msgConfig)
	pub1 := publisher.NewPublisher(domain, publisher1ID, configFolder, false, testMessenger)
	pub1.CreateNode(device1ID, device1Type)

	err := pub1.LoadRegisteredNodes()
	assert.NoErrorf(t, err, "Unable to load config from folder %s: %s", configFolder, err)
}

// TestAlias tests the use of alias in the inout discovery publication
func TestDiscoveryWithAlias(t *testing.T) {
	var testMessenger = messaging.NewDummyMessenger(msgConfig)
	pub1 := publisher.NewPublisher(domain, publisher1ID, configFolder, false, testMessenger)

	// update the node alias and test if node, input and outputs are published using their alias as nodeID
	pub1.Start()
	pub1.CreateNode(node1ID, types.NodeTypeUnknown)
	// pub1.CreateInput(node1ID, node1InputType, types.DefaultInputInstance, nil) // 4th [3]
	// pub1.CreateOutput(node1ID, node1Output1Type, types.DefaultOutputInstance)  // 4th [3]
	pub1.PublishUpdates()
	pub1.PublishNodeAlias(node1Addr, node1AliasID)

	// time.Sleep(1)
	// nodes, inputs and outputs must have been published using their alias
	// this should only affect domain discovered nodes, not registered nodes/inputs/outputs
	node := pub1.GetNodeByNodeID(node1AliasID)
	assert.NotNil(t, node, "Node not found using alias")

	pub1.Stop()
}

// TestReceiveInput tests receiving input control commands
func TestReceiveInput(t *testing.T) {
	var testMessenger = messaging.NewDummyMessenger(msgConfig)
	var node1Base = fmt.Sprintf("%s/%s/%s", domain, publisher1ID, node1ID)
	var node2Base = fmt.Sprintf("%s/%s/%s", domain, publisher2ID, "node2")
	var node1InputSetAddr = fmt.Sprintf("%s/%s/0/%s", node1Base, node1InputType, types.MessageTypeSet)
	var node2InputSetAddr = fmt.Sprintf("%s/%s/0/%s", node2Base, node1InputType, types.MessageTypeSet)

	// signMessages = false
	pub1 := publisher.NewPublisher(domain, publisher1ID, configFolder, false, testMessenger)

	pub1.Start()
	// update the node alias and see if its output is published with alias' as node id
	pub1.CreateInput(node1ID, types.InputTypeSwitch, types.DefaultInputInstance,
		func(input *types.InputDiscoveryMessage, sender string, value string) {
			logrus.Infof("Received message '%s' from sender %s", value, sender)
			pub1.UpdateOutputValue(node1ID, node1Output1Type, types.DefaultOutputInstance, value)
		})
	pub1.CreateNode(node1ID, types.NodeTypeUnknown)
	// pub1.CreateInput(node1ID, node1InputType, types.DefaultInputInstance, nil)
	pub1.CreateOutput(node1ID, node1Output1Type, types.DefaultOutputInstance)

	pub1.PublishUpdates()

	// process background messages
	// time.Sleep(time.Second * 1) // receive publications

	// test - Pass a set input command to the onreceive handler
	err := pub1.PublishSetInput(node1InputSetAddr, "true")
	assert.NoErrorf(t, err, "Publish failed: ", err)

	in1 := pub1.GetInputByAddress(node1InputAddr)
	assert.NotNilf(t, in1, "Input 1 not found on address %s", node1InputAddr)

	val := pub1.GetOutputValue(node1ID, node1Output1Type, types.DefaultOutputInstance)
	if !assert.NotNilf(t, val, "Unable to find output value for output %s", node1Output1Addr) {
		return
	}
	assert.Equal(t, "true", val.Value, "Input value didn't update the output")

	// error case - unknown publisher in input address should fail
	err = pub1.PublishSetInput(node2InputSetAddr, "true")
	assert.Errorf(t, err, "Publish should have failed as receiving publisher is unknown")

	pub1.Stop()
}

// TODO
// func TestPublishEvent(t *testing.T) {
// 	// setup
// 	var testMessenger = messaging.NewDummyMessenger(msgConfig)
// 	pub1 := publisher.NewPublisher(domain, publisher1ID, configFolder, false, testMessenger)
// 	publisher.PublishOutputEvent(node, pub1.reg)
// }

// run a bunch of facade commands with invalid arguments
func TestErrors(t *testing.T) {
	var testMessenger = messaging.NewDummyMessenger(msgConfig)
	pub1 := publisher.NewPublisher("", publisher1ID, configFolder, false, testMessenger)
	pub1.Address()
	pub1.CreateInput("fakeid", "faketype", types.DefaultInputInstance, nil)
	pub1.CreateInputFromFile("fakeid", "faketype", types.DefaultInputInstance, "fakepath", nil)
	pub1.CreateInputFromHTTP("fakeid", "faketype", types.DefaultInputInstance, "fakeurl", 0, nil)
	pub1.CreateInputFromOutput("fakeid", "faketype", types.DefaultInputInstance, "fakeaddr", nil)
	pub1.CreateNode("fakeid", types.NodeTypeAlarm)
	out1 := pub1.CreateOutput("fakeid", types.OutputTypeAlarm, types.DefaultOutputInstance)
	pub1.Domain()
	pub1.GetDomainInputs()
	pub1.GetDomainNodes()
	pub1.GetDomainOutputs()
	pub1.GetDomainPublishers()
	pub1.GetDomainInput("fakeaddr")
	pub1.GetDomainNode("fakeaddr")
	pub1.GetDomainOutput("fakeaddr")
	pub1.GetIdentity()
	pub1.GetIdentityKeys()
	pub1.GetInputByDevice("fakenode", "", "")
	pub1.GetInputs()
	pub1.GetNodeAttr("fakenode", "fakeattr")
	pub1.GetNodeByAddress("fakeaddr")
	pub1.GetNodeByDeviceID("fakenode")
	pub1.GetNodeByNodeID("fakenode")
	pub1.GetNodeConfigBool("fakeid", "fakeattr", false)
	pub1.GetNodeConfigFloat("fakeid", "fakeattr", 42.0)
	pub1.GetNodeConfigInt("fakeid", "fakeattr", 42)
	pub1.GetNodeConfigString("fakeid", "fakeattr", "fake")
	pub1.GetNodes()
	pub1.GetNodeStatus("fakeid", "fakeattr")
	pub1.GetOutput("fakenode", "", "")
	pub1.GetOutputs()
	pub1.GetOutputValue("fakeid", "faketype", "")
	pub1.MakeNodeDiscoveryAddress("fakeid")
	pub1.PublishNodeConfigure("fakeaddr", types.NodeAttrMap{})
	pub1.PublishRaw(out1, true, "value")
	pub1.SetSigningOnOff(true)
	pub1.UpdateNodeErrorStatus("fakeid", types.NodeRunStateError, "fake status")
	pub1.UpdateNodeAttr("fakeid", types.NodeAttrMap{})
	pub1.UpdateNodeConfig("fakeid", types.NodeAttrName, nil)
	pub1.UpdateNodeConfigValues("fakeid", types.NodeAttrMap{})
	pub1.UpdateNodeStatus("fakeid", types.NodeStatusMap{})
	pub1.UpdateOutputForecast("fakeid", []types.OutputValue{})
}
