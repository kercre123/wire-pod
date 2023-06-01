// ipc_manager.go handles messages between gateway and the following other processes:
//   - vic-engine: There are both a CLAD (EngineCladIpcManager) and a Protobuf (EngineProtoIpcManager)
//                 domain socket connecting gateway to the engine. The CLAD socket is currently in the
//                 process of being deprecated because it was designed as a temporary connection while
//                 messages were converted to Protobuf.
//   - vic-switchboard: There is a CLAD domain socket connecting gateway to switchboard. This socket is
//                      used to coordinate authentication between the two processes.
//   - vic-cloud: There is a CLAD domain socket connecting gateway to vic-cloud. This socket is used to
//                refresh the latest authentication tokens from vic-cloud. This socket is defined inside
//                the tokens.go file because the properties of that socket are a special case.
// To add a new connection, add the domain socket name (as defined by the server) to the list of consts
// below. And create a new IpcManager struct for that given socket. In main.go the connection should be
// Init-ed. Then it will be ready to use.
package main

import (
	"bytes"
	"encoding/binary"
	"math/rand"
	"path"
	"reflect"
	"sync"
	"time"

	"github.com/digital-dream-labs/vector-cloud/internal/ipc"

	gw_clad "github.com/digital-dream-labs/vector-cloud/internal/clad/gateway"
	extint "github.com/digital-dream-labs/vector-cloud/internal/proto/external_interface"

	"github.com/digital-dream-labs/vector-cloud/internal/log"

	"github.com/golang/protobuf/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
)

// The names of the domain sockets to which gateway will connect.
// They will be created in the location defined by config_{platform}.go for the appropriate
// platform.
const (
	cladDomainSocket        = "_engine_gateway_server_"
	protoDomainSocket       = "_engine_gateway_proto_server_"
	switchboardDomainSocket = "_switchboard_gateway_server_"
)

// IpcManager is a struct which handles synchronously sending and receiving messages from
// other processes via domain sockets. This acts like a base class for the other IpcManagers
// to prevent duplication of the Connect and Close functions.
//
// field connMutex: A mutex to prevent simultaneous reads and writes to the domain socket.
// field conn: The connection to the domain socket by which messages are passed between processes.
//
// Note: If there is a function that can be shared, use the engine manager.
//       The only reason I duplicated the functions so far was interfaces
//       were strangely causing the deleteListenerUnsafe to fail. - shawn 7/17/18
type IpcManager struct {
	connMutex sync.Mutex
	conn      ipc.Conn
}

// Connect establishes a connection to the domain socket of a given path.
// vic-gateway acts as the client in all of its domain socket connections since
// it is the part of the system closest to the outside world.
func (manager *IpcManager) Connect(path string, name string) {
	for {
		conn, err := ipc.NewUnixgramClient(path, name)
		if err != nil {
			log.Printf("Couldn't create sockets for %s & %s_%s - retrying: %s\n", path, path, name, err.Error())
			time.Sleep(5 * time.Second)
		} else {
			manager.conn = conn
			return
		}
	}
}

// Close tears down the connection to a domain socket.
func (manager *IpcManager) Close() error {
	return manager.conn.Close()
}

// EngineProtoIpcManager handles passing Protobuf messages between vic-gateway and vic-engine.
// field IpcManager: An anonymous field which basically acts like a subclass.
// field managerMutex: A mutex which prevents asynchronous reads and writes to the managedChannels map.
// field managedChannels: A mapping of messages to listening channels for the rpc handlers.
type EngineProtoIpcManager struct {
	IpcManager
	managerMutex    sync.RWMutex
	managedChannels map[string]([]chan extint.GatewayWrapper)
}

// Init sets up the channel manager and the domain socket connection
func (manager *EngineProtoIpcManager) Init() {
	manager.managedChannels = make(map[string]([]chan extint.GatewayWrapper))
	manager.Connect(path.Join(SocketPath, protoDomainSocket), "client")
}

// Write sends a Protobuf message to vic-engine. This will be handled by
// ProtoMessageHandler::ProcessMessages() in the engine C++ code.
// All Gateway messages get a ConnectionId, we create one here and return it.
func (manager *EngineProtoIpcManager) Write(msg *extint.GatewayWrapper) (int, uint64, error) {
	connectionID := rand.Uint64()
	count, err := manager.WriteWithID(msg, connectionID)
	return count, connectionID, err
}

// Write sends a Protobuf message to vic-engine. This will be handled by
// ProtoMessageHandler::ProcessMessages() in the engine C++ code.
func (manager *EngineProtoIpcManager) WriteWithID(msg *extint.GatewayWrapper, connID uint64) (int, error) {
	var err error
	var buf bytes.Buffer

	msg.ConnectionId = connID

	if msg == nil {
		return -1, grpc.Errorf(codes.InvalidArgument, "Unable to parse request")
	}
	if err = binary.Write(&buf, binary.LittleEndian, uint16(proto.Size(msg))); err != nil {
		return -1, grpc.Errorf(codes.Internal, err.Error())
	}
	data, err := proto.Marshal(msg)
	if err != nil {
		return -1, grpc.Errorf(codes.Internal, err.Error())
	}
	if err = binary.Write(&buf, binary.LittleEndian, data); err != nil {
		return -1, grpc.Errorf(codes.Internal, err.Error())
	}

	manager.connMutex.Lock()
	defer manager.connMutex.Unlock()
	if logMessageContent {
		log.Printf("%T: writing '%#v' Proto message to Engine\n", *manager, msg)
	}
	return manager.conn.Write(buf.Bytes())
}

// deleteListenerUnsafe is "unsafe" in that it requires the calling function to have already acquired the
// appropriate lock. Otherwise, there is a chance of simultaneous map access.
func (manager *EngineProtoIpcManager) deleteListenerUnsafe(listener chan extint.GatewayWrapper, tag string) {
	chanSlice := manager.managedChannels[tag]
	for idx, v := range chanSlice {
		if v == listener {
			chanSlice[idx] = chanSlice[len(chanSlice)-1]
			manager.managedChannels[tag] = chanSlice[:len(chanSlice)-1]
			manager.SafeClose(listener)
			break
		}
		if len(chanSlice)-1 == idx {
			log.Println("Warning: failed to remove listener:", listener, "from array:", chanSlice)
		}
	}
	if len(manager.managedChannels[tag]) == 0 {
		delete(manager.managedChannels, tag)
	}
}

// deleteListenerCallback provides a callback which may be invoked to remove the channel from the listeners.
func (manager *EngineProtoIpcManager) deleteListenerCallback(listener chan extint.GatewayWrapper, tag string) func() {
	return func() {
		manager.managerMutex.Lock()
		defer manager.managerMutex.Unlock()
		manager.deleteListenerUnsafe(listener, tag)
	}
}

// CreateChannel establishes a chan by which the ipc manager will pass the requested message type.
func (manager *EngineProtoIpcManager) CreateChannel(tag interface{}, numChannels int) (func(), chan extint.GatewayWrapper) {
	result := make(chan extint.GatewayWrapper, numChannels)
	reflectedType := reflect.TypeOf(tag).String()
	if logVerbose {
		log.Println("Listening for", reflectedType)
	}
	manager.managerMutex.Lock()
	defer manager.managerMutex.Unlock()
	slice := manager.managedChannels[reflectedType]
	if slice == nil {
		slice = make([]chan extint.GatewayWrapper, 0)
	}
	manager.managedChannels[reflectedType] = append(slice, result)
	return manager.deleteListenerCallback(result, reflectedType), result
}

// CreateUniqueChannel establishes a chan by which the ipc manager will pass the requested message type as long as there isn't already one open.
// This is useful in cases where we want to prevent spamming of the engine. See UpdateSettings in message_handler.go.
func (manager *EngineProtoIpcManager) CreateUniqueChannel(tag interface{}, numChannels int) (func(), chan extint.GatewayWrapper, bool) {
	reflectedType := reflect.TypeOf(tag).String()
	manager.managerMutex.RLock()
	_, ok := manager.managedChannels[reflectedType]
	manager.managerMutex.RUnlock()
	if ok {
		return nil, nil, false
	}

	f, c := manager.CreateChannel(tag, numChannels)
	return f, c, true
}

// SafeClose closes a channel if possible.
func (manager *EngineProtoIpcManager) SafeClose(listener chan extint.GatewayWrapper) {
	select {
	case _, ok := <-listener:
		if ok {
			close(listener)
		}
	default:
		close(listener)
	}
}

// SendToListeners propagates messages to all waiting listener channels.
func (manager *EngineProtoIpcManager) SendToListeners(tag string, msg extint.GatewayWrapper) {
	markedForDelete := make(chan chan extint.GatewayWrapper, 5)
	defer func() {
		close(markedForDelete)
		if len(markedForDelete) != 0 {
			manager.managerMutex.Lock()
			defer manager.managerMutex.Unlock()
			for listener := range markedForDelete {
				manager.deleteListenerUnsafe(listener, tag)
			}
		}
	}()
	manager.managerMutex.RLock()
	defer manager.managerMutex.RUnlock()
	chanList, ok := manager.managedChannels[tag]
	if !ok {
		return // No listeners for message
	}
	if logVerbose {
		log.Printf("Sending %s to listeners\n", tag)
	}
	var wg sync.WaitGroup
	for idx, listener := range chanList {
		wg.Add(1)
		go func(idx int, listener chan extint.GatewayWrapper, msg extint.GatewayWrapper) {
			defer wg.Done()
			select {
			case listener <- msg:
				if logVerbose {
					log.Printf("Sent to listener #%d: %s\n", idx, tag)
				}
			case <-time.After(250 * time.Millisecond):
				log.Errorf("EngineProtoIpcManager.SendToListeners: Failed to send message %s for listener #%d. There might be a problem with the channel.\n", tag, idx)
				markedForDelete <- listener
			}
		}(idx, listener, msg)
	}
	wg.Wait()
}

// ProcessMessages loops through incoming messages on the ipc channel.
func (manager *EngineProtoIpcManager) ProcessMessages() {
	var msg extint.GatewayWrapper
	var b, block []byte
	for {
		msg.Reset()
		block = manager.conn.ReadBlock()
		if block == nil {
			log.Errorln("EngineProtoIpcManager.ProcessMessages.Fatal: engine socket returned empty message")
			return
		} else if len(block) < 2 {
			log.Errorln("EngineProtoIpcManager.ProcessMessages.Error: engine socket message too small")
			continue
		}
		b = block[2:]

		if err := proto.Unmarshal(b, &msg); err != nil {
			log.Errorln("EngineProtoIpcManager.DecodingError (", err, "):", b)
			continue
		}
		tag := reflect.TypeOf(msg.OneofMessageType).String()
		manager.SendToListeners(tag, msg)
	}
}

// SwitchboardIpcManager handles passing CLAD messages between vic-gateway and vic-switchboard.
// field IpcManager: An anonymous field which basically acts like a subclass.
// field managerMutex: A mutex which prevents asynchronous reads and writes to the managedChannels map.
// field managedChannels: A mapping of messages to listening channels for the rpc handlers.
type SwitchboardIpcManager struct {
	IpcManager
	managerMutex    sync.RWMutex
	managedChannels map[gw_clad.SwitchboardResponseTag]([]chan gw_clad.SwitchboardResponse)
}

// Init sets up the channel manager and the domain socket connection
func (manager *SwitchboardIpcManager) Init() {
	manager.managedChannels = make(map[gw_clad.SwitchboardResponseTag]([]chan gw_clad.SwitchboardResponse))
	manager.Connect(path.Join(SocketPath, switchboardDomainSocket), "client")
}

func (manager *SwitchboardIpcManager) handleSwitchboardMessages(msg *gw_clad.SwitchboardResponse) {
	switch msg.Tag() {
	case gw_clad.SwitchboardResponseTag_SdkProxyRequest:
		request := msg.GetSdkProxyRequest()
		go func() {
			manager.Write(
				gw_clad.NewSwitchboardRequestWithSdkProxyResponse(
					bleProxy.handle(request),
				),
			)
		}()
	case gw_clad.SwitchboardResponseTag_ExternalConnectionRequest:
		manager.Write(gw_clad.NewSwitchboardRequestWithExternalConnectionResponse(
			&gw_clad.ExternalConnectionResponse{
				IsConnected:  len(connectionId) != 0,
				ConnectionId: connectionId,
			},
		))
	case gw_clad.SwitchboardResponseTag_ClientGuidRefreshRequest:
		response := make(chan struct{})
		tokenManager.ForceUpdate(response)
		<-response
		manager.Write(gw_clad.NewSwitchboardRequestWithClientGuidRefreshResponse(
			&gw_clad.ClientGuidRefreshResponse{},
		))
	}
}

// ProcessMessages loops through incoming messages on the ipc channel.
func (manager *SwitchboardIpcManager) ProcessMessages() {
	var msg gw_clad.SwitchboardResponse
	var b, block []byte
	var buf bytes.Buffer
	for {
		buf.Reset()
		block = manager.conn.ReadBlock()
		if block == nil {
			log.Errorln("SwitchboardIpcManager.ProcessMessages.Fatal: switchboard socket returned empty message")
			return
		} else if len(block) < 2 {
			log.Errorln("SwitchboardIpcManager.ProcessMessages.Error: switchboard socket message too small")
			continue
		}
		b = block[2:]
		buf.Write(b)
		if err := msg.Unpack(&buf); err != nil {
			log.Errorln("SwitchboardIpcManager.DecodingError (", err, "):", b)
			continue
		}

		manager.handleSwitchboardMessages(&msg)
		manager.SendToListeners(msg)
	}
}

// Write sends a CLAD message to vic-switchboard.
func (manager *SwitchboardIpcManager) Write(msg *gw_clad.SwitchboardRequest) (int, error) {
	var err error
	var buf bytes.Buffer
	if msg == nil {
		return -1, grpc.Errorf(codes.InvalidArgument, "Unable to parse request")
	}
	if err = binary.Write(&buf, binary.LittleEndian, uint16(msg.Size())); err != nil {
		return -1, grpc.Errorf(codes.Internal, err.Error())
	}
	if err = msg.Pack(&buf); err != nil {
		return -1, grpc.Errorf(codes.Internal, err.Error())
	}

	manager.connMutex.Lock()
	defer manager.connMutex.Unlock()
	if logMessageContent {
		log.Printf("%T: writing '%#v' message to Switchboard\n", *manager, *msg)
	}
	return manager.conn.Write(buf.Bytes())
}

// deleteListenerUnsafe is "unsafe" in that it requires the calling function to have already acquired the
// appropriate lock. Otherwise, there is a chance of simultaneous map access.
func (manager *SwitchboardIpcManager) deleteListenerUnsafe(listener chan gw_clad.SwitchboardResponse, tag gw_clad.SwitchboardResponseTag) {
	chanSlice := manager.managedChannels[tag]
	for idx, v := range chanSlice {
		if v == listener {
			chanSlice[idx] = chanSlice[len(chanSlice)-1]
			manager.managedChannels[tag] = chanSlice[:len(chanSlice)-1]
			manager.SafeClose(listener)
			break
		}
		if len(chanSlice)-1 == idx {
			log.Println("Warning: failed to remove listener:", listener, "from array:", chanSlice)
		}
	}
	if len(manager.managedChannels[tag]) == 0 {
		delete(manager.managedChannels, tag)
	}
}

// deleteListenerCallback provides a callback which may be invoked to remove the channel from the listeners.
func (manager *SwitchboardIpcManager) deleteListenerCallback(listener chan gw_clad.SwitchboardResponse, tag gw_clad.SwitchboardResponseTag) func() {
	return func() {
		manager.managerMutex.Lock()
		defer manager.managerMutex.Unlock()
		manager.deleteListenerUnsafe(listener, tag)
	}
}

// CreateChannel establishes a chan by which the ipc manager will pass the requested message type.
func (manager *SwitchboardIpcManager) CreateChannel(tag gw_clad.SwitchboardResponseTag, numChannels int) (func(), chan gw_clad.SwitchboardResponse) {
	result := make(chan gw_clad.SwitchboardResponse, numChannels)
	if logVerbose {
		log.Printf("Listening for %+v\n", tag)
	}
	manager.managerMutex.Lock()
	defer manager.managerMutex.Unlock()
	slice := manager.managedChannels[tag]
	if slice == nil {
		slice = make([]chan gw_clad.SwitchboardResponse, 0)
	}
	manager.managedChannels[tag] = append(slice, result)
	return manager.deleteListenerCallback(result, tag), result
}

// SafeClose closes a channel if possible.
func (manager *SwitchboardIpcManager) SafeClose(listener chan gw_clad.SwitchboardResponse) {
	select {
	case _, ok := <-listener:
		if ok {
			close(listener)
		}
	default:
		close(listener)
	}
}

// SendToListeners propagates messages to all waiting listener channels.
func (manager *SwitchboardIpcManager) SendToListeners(msg gw_clad.SwitchboardResponse) {
	tag := msg.Tag()
	markedForDelete := make(chan chan gw_clad.SwitchboardResponse, 5)
	defer func() {
		close(markedForDelete)
		if len(markedForDelete) != 0 {
			manager.managerMutex.Lock()
			defer manager.managerMutex.Unlock()
			for listener := range markedForDelete {
				manager.deleteListenerUnsafe(listener, tag)
			}
		}
	}()
	manager.managerMutex.RLock()
	defer manager.managerMutex.RUnlock()
	chanList, ok := manager.managedChannels[tag]
	if !ok {
		return // No listeners for message
	}
	if logVerbose {
		log.Printf("Sending %s to listeners\n", tag)
	}
	var wg sync.WaitGroup
	for idx, listener := range chanList {
		wg.Add(1)
		go func(idx int, listener chan gw_clad.SwitchboardResponse, msg gw_clad.SwitchboardResponse) {
			defer wg.Done()
			select {
			case listener <- msg:
				if logVerbose {
					log.Printf("Sent to listener #%d: %s\n", idx, tag)
				}
			case <-time.After(250 * time.Millisecond):
				log.Errorf("SwitchboardIpcManager.SendToListeners: Failed to send message %s for listener #%d. There might be a problem with the channel.\n", tag, idx)
				markedForDelete <- listener
			}
		}(idx, listener, msg)
	}
	wg.Wait()
}

// TODO: Remove CLAD manager once it's no longer needed

// EngineCladIpcManager handles passing CLAD messages between vic-gateway and vic-engine.
// field IpcManager: An anonymous field which basically acts like a subclass.
// field managerMutex: A mutex which prevents asynchronous reads and writes to the managedChannels map.
// field managedChannels: A mapping of messages to listening channels for the rpc handlers.
type EngineCladIpcManager struct {
	IpcManager
	managerMutex    sync.RWMutex
	managedChannels map[gw_clad.MessageRobotToExternalTag]([]chan gw_clad.MessageRobotToExternal)
}

// Init sets up the channel manager and the domain socket connection
func (manager *EngineCladIpcManager) Init() {
	manager.managedChannels = make(map[gw_clad.MessageRobotToExternalTag]([]chan gw_clad.MessageRobotToExternal))
	manager.Connect(path.Join(SocketPath, cladDomainSocket), "client")
}

// Write sends a CLAD message to vic-engine. This will be handled by
// UiMessageHandler::ProcessMessages() in the engine C++ code.
func (manager *EngineCladIpcManager) Write(msg *gw_clad.MessageExternalToRobot) (int, error) {
	var err error
	var buf bytes.Buffer
	if msg == nil {
		return -1, grpc.Errorf(codes.InvalidArgument, "Unable to parse request")
	}
	if err = binary.Write(&buf, binary.LittleEndian, uint16(msg.Size())); err != nil {
		return -1, grpc.Errorf(codes.Internal, err.Error())
	}
	if err = msg.Pack(&buf); err != nil {
		return -1, grpc.Errorf(codes.Internal, err.Error())
	}

	manager.connMutex.Lock()
	defer manager.connMutex.Unlock()
	if logMessageContent {
		log.Printf("%T: writing '%#v' CLAD message to Engine\n", *manager, *msg)
	}
	return manager.conn.Write(buf.Bytes())
}

// deleteListenerUnsafe is "unsafe" in that it requires the calling function to have already acquired the
// appropriate lock. Otherwise, there is a chance of simultaneous map access.
func (manager *EngineCladIpcManager) deleteListenerUnsafe(listener chan gw_clad.MessageRobotToExternal, tag gw_clad.MessageRobotToExternalTag) {
	chanSlice := manager.managedChannels[tag]
	for idx, v := range chanSlice {
		if v == listener {
			chanSlice[idx] = chanSlice[len(chanSlice)-1]
			manager.managedChannels[tag] = chanSlice[:len(chanSlice)-1]
			manager.SafeClose(listener)
			break
		}
		if len(chanSlice)-1 == idx {
			log.Println("Warning: failed to remove listener:", listener, "from array:", chanSlice)
		}
	}
	if len(manager.managedChannels[tag]) == 0 {
		delete(manager.managedChannels, tag)
	}
}

// deleteListenerCallback provides a callback which may be invoked to remove the channel from the listeners.
func (manager *EngineCladIpcManager) deleteListenerCallback(listener chan gw_clad.MessageRobotToExternal, tag gw_clad.MessageRobotToExternalTag) func() {
	return func() {
		manager.managerMutex.Lock()
		defer manager.managerMutex.Unlock()
		manager.deleteListenerUnsafe(listener, tag)
	}
}

// CreateChannel establishes a chan by which the ipc manager will pass the requested message type.
func (manager *EngineCladIpcManager) CreateChannel(tag gw_clad.MessageRobotToExternalTag, numChannels int) (func(), chan gw_clad.MessageRobotToExternal) {
	result := make(chan gw_clad.MessageRobotToExternal, numChannels)
	if logVerbose {
		log.Printf("Listening for %+v\n", tag)
	}
	manager.managerMutex.Lock()
	defer manager.managerMutex.Unlock()
	slice := manager.managedChannels[tag]
	if slice == nil {
		slice = make([]chan gw_clad.MessageRobotToExternal, 0)
	}
	manager.managedChannels[tag] = append(slice, result)
	return manager.deleteListenerCallback(result, tag), result
}

// SafeClose closes a channel if possible.
func (manager *EngineCladIpcManager) SafeClose(listener chan gw_clad.MessageRobotToExternal) {
	select {
	case _, ok := <-listener:
		if ok {
			close(listener)
		}
	default:
		close(listener)
	}
}

// SendToListeners propagates messages to all waiting listener channels.
func (manager *EngineCladIpcManager) SendToListeners(msg gw_clad.MessageRobotToExternal) {
	tag := msg.Tag()
	markedForDelete := make(chan chan gw_clad.MessageRobotToExternal, 5)
	defer func() {
		close(markedForDelete)
		if len(markedForDelete) != 0 {
			manager.managerMutex.Lock()
			defer manager.managerMutex.Unlock()
			for listener := range markedForDelete {
				manager.deleteListenerUnsafe(listener, tag)
			}
		}
	}()
	manager.managerMutex.RLock()
	defer manager.managerMutex.RUnlock()
	chanList, ok := manager.managedChannels[tag]
	if !ok {
		return // No listeners for message
	}
	if logVerbose {
		log.Printf("Sending %s to listeners\n", tag)
	}
	var wg sync.WaitGroup
	for idx, listener := range chanList {
		wg.Add(1)
		go func(idx int, listener chan gw_clad.MessageRobotToExternal, msg gw_clad.MessageRobotToExternal) {
			defer wg.Done()
			select {
			case listener <- msg:
				if logVerbose {
					log.Printf("Sent to listener #%d: %s\n", idx, tag)
				}
			case <-time.After(250 * time.Millisecond):
				log.Errorf("EngineCladIpcManager.SendToListeners: Failed to send message %s for listener #%d. There might be a problem with the channel.\n", tag, idx)
				markedForDelete <- listener
			}
		}(idx, listener, msg)
	}
	wg.Wait()
}

// SendEventToChannel is a temporary function to more easily turn CLAD messages into Protobuf events.
func (manager *EngineCladIpcManager) SendEventToChannel(event *extint.Event) {
	tag := reflect.TypeOf(&extint.GatewayWrapper_Event{}).String()
	msg := extint.GatewayWrapper{
		OneofMessageType: &extint.GatewayWrapper_Event{
			// TODO: Convert all events into proto events
			Event: event,
		},
	}
	engineProtoManager.SendToListeners(tag, msg)
}

// ProcessMessages loops through incoming messages on the ipc channel.
// Note: this will ignore unparsable messages because there are more
// clad messages sent from engine than understood by gateway.
func (manager *EngineCladIpcManager) ProcessMessages() {
	var msg gw_clad.MessageRobotToExternal
	var b, block []byte
	var buf bytes.Buffer
	for {
		buf.Reset()
		block = manager.conn.ReadBlock()
		if block == nil {
			log.Errorln("EngineCladIpcManager.ProcessMessages.Fatal: engine socket returned empty message")
			return
		} else if len(block) < 2 {
			log.Errorln("EngineCladIpcManager.ProcessMessages.Error: engine socket message too small")
			continue
		}
		b = block[2:]
		buf.Write(b)
		if err := msg.Unpack(&buf); err != nil {
			// Intentionally ignoring errors for unknown messages
			// TODO: treat this as an error condition once VIC-3186 is completed
			continue
		}

		// TODO: Refactor now that RobotObservedFace and RobotChangedObservedFaceID have been added to events
		switch msg.Tag() {
		case gw_clad.MessageRobotToExternalTag_Event:
			event := CladEventToProto(msg.GetEvent())
			manager.SendEventToChannel(event)
			// @TODO: Convert all face events to proto VIC-4643
		case gw_clad.MessageRobotToExternalTag_RobotObservedFace:
			event := &extint.Event{
				EventType: &extint.Event_RobotObservedFace{
					RobotObservedFace: CladRobotObservedFaceToProto(msg.GetRobotObservedFace()),
				},
			}
			manager.SendEventToChannel(event)
		case gw_clad.MessageRobotToExternalTag_RobotChangedObservedFaceID:
			event := &extint.Event{
				EventType: &extint.Event_RobotChangedObservedFaceId{
					RobotChangedObservedFaceId: CladRobotChangedObservedFaceIDToProto(msg.GetRobotChangedObservedFaceID()),
				},
			}
			manager.SendEventToChannel(event)
			// @TODO: Convert all object events to proto VIC-4643
		case gw_clad.MessageRobotToExternalTag_ObjectAvailable:
			event := &extint.Event{
				EventType: &extint.Event_ObjectEvent{
					ObjectEvent: &extint.ObjectEvent{
						ObjectEventType: &extint.ObjectEvent_ObjectAvailable{
							ObjectAvailable: CladObjectAvailableToProto(msg.GetObjectAvailable()),
						},
					},
				},
			}
			manager.SendEventToChannel(event)
		case gw_clad.MessageRobotToExternalTag_ObjectConnectionState:
			event := &extint.Event{
				EventType: &extint.Event_ObjectEvent{
					ObjectEvent: &extint.ObjectEvent{
						ObjectEventType: &extint.ObjectEvent_ObjectConnectionState{
							ObjectConnectionState: CladObjectConnectionStateToProto(msg.GetObjectConnectionState()),
						},
					},
				},
			}
			manager.SendEventToChannel(event)
		case gw_clad.MessageRobotToExternalTag_ObjectMoved:
			event := &extint.Event{
				EventType: &extint.Event_ObjectEvent{
					ObjectEvent: &extint.ObjectEvent{
						ObjectEventType: &extint.ObjectEvent_ObjectMoved{
							ObjectMoved: CladObjectMovedToProto(msg.GetObjectMoved()),
						},
					},
				},
			}
			manager.SendEventToChannel(event)
		case gw_clad.MessageRobotToExternalTag_ObjectStoppedMoving:
			event := &extint.Event{
				EventType: &extint.Event_ObjectEvent{
					ObjectEvent: &extint.ObjectEvent{
						ObjectEventType: &extint.ObjectEvent_ObjectStoppedMoving{
							ObjectStoppedMoving: CladObjectStoppedMovingToProto(msg.GetObjectStoppedMoving()),
						},
					},
				},
			}
			manager.SendEventToChannel(event)
		case gw_clad.MessageRobotToExternalTag_ObjectUpAxisChanged:
			event := &extint.Event{
				EventType: &extint.Event_ObjectEvent{
					ObjectEvent: &extint.ObjectEvent{
						ObjectEventType: &extint.ObjectEvent_ObjectUpAxisChanged{
							ObjectUpAxisChanged: CladObjectUpAxisChangedToProto(msg.GetObjectUpAxisChanged()),
						},
					},
				},
			}
			manager.SendEventToChannel(event)
		case gw_clad.MessageRobotToExternalTag_ObjectTapped:
			event := &extint.Event{
				EventType: &extint.Event_ObjectEvent{
					ObjectEvent: &extint.ObjectEvent{
						ObjectEventType: &extint.ObjectEvent_ObjectTapped{
							ObjectTapped: CladObjectTappedToProto(msg.GetObjectTapped()),
						},
					},
				},
			}
			manager.SendEventToChannel(event)
		case gw_clad.MessageRobotToExternalTag_RobotObservedObject:
			event := &extint.Event{
				EventType: &extint.Event_ObjectEvent{
					ObjectEvent: &extint.ObjectEvent{
						ObjectEventType: &extint.ObjectEvent_RobotObservedObject{
							RobotObservedObject: CladRobotObservedObjectToProto(msg.GetRobotObservedObject()),
						},
					},
				},
			}
			manager.SendEventToChannel(event)

		default:
			manager.SendToListeners(msg)
		}
	}
}
