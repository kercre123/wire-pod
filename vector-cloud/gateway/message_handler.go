package main

import (
	"encoding/binary"
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"os/exec"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	cloud_clad "github.com/digital-dream-labs/vector-cloud/internal/clad/cloud"
	gw_clad "github.com/digital-dream-labs/vector-cloud/internal/clad/gateway"
	extint "github.com/digital-dream-labs/vector-cloud/internal/proto/external_interface"

	"github.com/digital-dream-labs/vector-cloud/internal/log"

	"github.com/golang/protobuf/proto"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
)

const faceImagePixelsPerChunk = 600
const endOfAnimationList = "EndOfListAnimationsResponses"

var (
	connectionIdLock    sync.Mutex
	connectionId        string
	statusStreamRunning bool
	lastProgress        int64
	lastExpected        int64
)

// TODO: we should find a way to auto-generate the equivalent of this function as part of clad or protoc
func ProtoMoveHeadToClad(msg *extint.MoveHeadRequest) *gw_clad.MessageExternalToRobot {
	return gw_clad.NewMessageExternalToRobotWithMoveHead(&gw_clad.MoveHead{
		SpeedRadPerSec: msg.SpeedRadPerSec,
	})
}

// TODO: we should find a way to auto-generate the equivalent of this function as part of clad or protoc
func ProtoMoveLiftToClad(msg *extint.MoveLiftRequest) *gw_clad.MessageExternalToRobot {
	return gw_clad.NewMessageExternalToRobotWithMoveLift(&gw_clad.MoveLift{
		SpeedRadPerSec: msg.SpeedRadPerSec,
	})
}

// TODO: we should find a way to auto-generate the equivalent of this function as part of clad or protoc
func FaceImageChunkToClad(faceData [faceImagePixelsPerChunk]uint16, pixelCount uint16, chunkIndex uint8, chunkCount uint8, durationMs uint32, interruptRunning bool) *gw_clad.MessageExternalToRobot {
	return gw_clad.NewMessageExternalToRobotWithDisplayFaceImageRGBChunk(&gw_clad.DisplayFaceImageRGBChunk{
		FaceData:         faceData,
		NumPixels:        pixelCount,
		ChunkIndex:       chunkIndex,
		NumChunks:        chunkCount,
		DurationMs:       durationMs,
		InterruptRunning: interruptRunning,
	})
}

func ProtoAppIntentToClad(msg *extint.AppIntentRequest) *gw_clad.MessageExternalToRobot {
	return gw_clad.NewMessageExternalToRobotWithAppIntent(&gw_clad.AppIntent{
		Param:  msg.Param,
		Intent: msg.Intent,
	})
}

func ProtoRequestEnrolledNamesToClad(msg *extint.RequestEnrolledNamesRequest) *gw_clad.MessageExternalToRobot {
	return gw_clad.NewMessageExternalToRobotWithRequestEnrolledNames(&gw_clad.RequestEnrolledNames{})
}

func ProtoCancelFaceEnrollmentToClad(msg *extint.CancelFaceEnrollmentRequest) *gw_clad.MessageExternalToRobot {
	return gw_clad.NewMessageExternalToRobotWithCancelFaceEnrollment(&gw_clad.CancelFaceEnrollment{})
}

func ProtoUpdateEnrolledFaceByIDToClad(msg *extint.UpdateEnrolledFaceByIDRequest) *gw_clad.MessageExternalToRobot {
	return gw_clad.NewMessageExternalToRobotWithUpdateEnrolledFaceByID(&gw_clad.UpdateEnrolledFaceByID{
		FaceID:  msg.FaceId,
		OldName: msg.OldName,
		NewName: msg.NewName,
	})
}

func ProtoEraseEnrolledFaceByIDToClad(msg *extint.EraseEnrolledFaceByIDRequest) *gw_clad.MessageExternalToRobot {
	return gw_clad.NewMessageExternalToRobotWithEraseEnrolledFaceByID(&gw_clad.EraseEnrolledFaceByID{
		FaceID: msg.FaceId,
	})
}

func ProtoEraseAllEnrolledFacesToClad(msg *extint.EraseAllEnrolledFacesRequest) *gw_clad.MessageExternalToRobot {
	return gw_clad.NewMessageExternalToRobotWithEraseAllEnrolledFaces(&gw_clad.EraseAllEnrolledFaces{})
}

func ProtoPoseToClad(msg *extint.PoseStruct) *gw_clad.PoseStruct3d {
	return &gw_clad.PoseStruct3d{
		X:        msg.X,
		Y:        msg.Y,
		Z:        msg.Z,
		Q0:       msg.Q0,
		Q1:       msg.Q1,
		Q2:       msg.Q2,
		Q3:       msg.Q3,
		OriginID: msg.OriginId,
	}
}

func ProtoCreateFixedCustomObjectToClad(msg *extint.CreateFixedCustomObjectRequest) *gw_clad.MessageExternalToRobot {
	return gw_clad.NewMessageExternalToRobotWithCreateFixedCustomObject(&gw_clad.CreateFixedCustomObject{
		Pose:    *ProtoPoseToClad(msg.Pose),
		XSizeMm: msg.XSizeMm,
		YSizeMm: msg.YSizeMm,
		ZSizeMm: msg.ZSizeMm,
	})
}

func ProtoDefineCustomBoxToClad(msg *extint.DefineCustomObjectRequest, def *extint.CustomBoxDefinition) *gw_clad.MessageExternalToRobot {
	// Convert from the proto defined CustomObject enum to the more general clad ObjectType enum space
	object_type := gw_clad.ObjectType(int(msg.CustomType) - int(extint.CustomType_CUSTOM_TYPE_00) + int(gw_clad.ObjectType_CustomType00))

	return gw_clad.NewMessageExternalToRobotWithDefineCustomBox(&gw_clad.DefineCustomBox{
		CustomType:     object_type,
		MarkerFront:    gw_clad.CustomObjectMarker(def.MarkerFront - 1),
		MarkerBack:     gw_clad.CustomObjectMarker(def.MarkerBack - 1),
		MarkerTop:      gw_clad.CustomObjectMarker(def.MarkerTop - 1),
		MarkerBottom:   gw_clad.CustomObjectMarker(def.MarkerBottom - 1),
		MarkerLeft:     gw_clad.CustomObjectMarker(def.MarkerLeft - 1),
		MarkerRight:    gw_clad.CustomObjectMarker(def.MarkerRight - 1),
		XSizeMm:        def.XSizeMm,
		YSizeMm:        def.YSizeMm,
		ZSizeMm:        def.ZSizeMm,
		MarkerWidthMm:  def.MarkerWidthMm,
		MarkerHeightMm: def.MarkerHeightMm,
		IsUnique:       msg.IsUnique,
	})
}

func ProtoDefineCustomCubeToClad(msg *extint.DefineCustomObjectRequest, def *extint.CustomCubeDefinition) *gw_clad.MessageExternalToRobot {
	// Convert from the proto defined CustomObject enum to the more general clad ObjectType enum space
	object_type := gw_clad.ObjectType(int(msg.CustomType) - int(extint.CustomType_CUSTOM_TYPE_00) + int(gw_clad.ObjectType_CustomType00))

	return gw_clad.NewMessageExternalToRobotWithDefineCustomCube(&gw_clad.DefineCustomCube{
		CustomType:     object_type,
		Marker:         gw_clad.CustomObjectMarker(def.Marker - 1),
		SizeMm:         def.SizeMm,
		MarkerWidthMm:  def.MarkerWidthMm,
		MarkerHeightMm: def.MarkerHeightMm,
		IsUnique:       msg.IsUnique,
	})
}

func ProtoDefineCustomWallToClad(msg *extint.DefineCustomObjectRequest, def *extint.CustomWallDefinition) *gw_clad.MessageExternalToRobot {
	// Convert from the proto defined CustomObject enum to the more general clad ObjectType enum space
	object_type := gw_clad.ObjectType(int(msg.CustomType) - int(extint.CustomType_CUSTOM_TYPE_00) + int(gw_clad.ObjectType_CustomType00))

	return gw_clad.NewMessageExternalToRobotWithDefineCustomWall(&gw_clad.DefineCustomWall{
		CustomType:     object_type,
		Marker:         gw_clad.CustomObjectMarker(def.Marker - 1),
		WidthMm:        def.WidthMm,
		HeightMm:       def.HeightMm,
		MarkerWidthMm:  def.MarkerWidthMm,
		MarkerHeightMm: def.MarkerHeightMm,
		IsUnique:       msg.IsUnique,
	})
}

func SliceToArray(msg []uint32) [3]uint32 {
	var arr [3]uint32
	copy(arr[:], msg)
	return arr
}

func CladCladRectToProto(msg *gw_clad.CladRect) *extint.CladRect {
	return &extint.CladRect{
		XTopLeft: msg.XTopLeft,
		YTopLeft: msg.YTopLeft,
		Width:    msg.Width,
		Height:   msg.Height,
	}
}

func CladCladPointsToProto(msg []gw_clad.CladPoint2d) []*extint.CladPoint {
	var points []*extint.CladPoint
	for _, point := range msg {
		points = append(points, &extint.CladPoint{X: point.X, Y: point.Y})
	}
	return points
}

func CladExpressionValuesToProto(msg []uint8) []uint32 {
	var expression_values []uint32
	for _, val := range msg {
		expression_values = append(expression_values, uint32(val))
	}
	return expression_values
}

func CladRobotObservedFaceToProto(msg *gw_clad.RobotObservedFace) *extint.RobotObservedFace {
	// BlinkAmount, Gaze and SmileAmount are not exposed to the SDK
	return &extint.RobotObservedFace{
		FaceId:    msg.FaceID,
		Timestamp: msg.Timestamp,
		Pose:      CladPoseToProto(&msg.Pose),
		ImgRect:   CladCladRectToProto(&msg.ImgRect),
		Name:      msg.Name,

		Expression: extint.FacialExpression(msg.Expression + 1), // protobuf enums have a 0 start value

		// Individual expression values histogram, sums to 100 (Exception: all zero if expressio: msg.
		ExpressionValues: CladExpressionValuesToProto(msg.ExpressionValues[:]),

		// Face landmarks
		LeftEye:  CladCladPointsToProto(msg.LeftEye),
		RightEye: CladCladPointsToProto(msg.RightEye),
		Nose:     CladCladPointsToProto(msg.Nose),
		Mouth:    CladCladPointsToProto(msg.Mouth),
	}
}

func CladRobotChangedObservedFaceIDToProto(msg *gw_clad.RobotChangedObservedFaceID) *extint.RobotChangedObservedFaceID {
	return &extint.RobotChangedObservedFaceID{
		OldId: msg.OldID,
		NewId: msg.NewID,
	}
}

func CladPoseToProto(msg *gw_clad.PoseStruct3d) *extint.PoseStruct {
	return &extint.PoseStruct{
		X:        msg.X,
		Y:        msg.Y,
		Z:        msg.Z,
		Q0:       msg.Q0,
		Q1:       msg.Q1,
		Q2:       msg.Q2,
		Q3:       msg.Q3,
		OriginId: msg.OriginID,
	}
}

func CladEventToProto(msg *gw_clad.Event) *extint.Event {
	switch tag := msg.Tag(); tag {
	// Event is currently unused in CLAD, but if you start
	// using it again, replace [MessageName] with your msg name
	// case gw_clad.EventTag_[MessageName]:
	// 	return &extint.Event{
	// 		EventType: &extint.Event_[MessageName]{
	// 			Clad[MessageName]ToProto(msg.Get[MessageName]()),
	// 		},
	// 	}
	case gw_clad.EventTag_INVALID:
		log.Println(tag, "tag is invalid")
		return nil
	default:
		log.Println(tag, "tag is not yet implemented")
		return nil
	}
}

func CladObjectConnectionStateToProto(msg *gw_clad.ObjectConnectionState) *extint.ObjectConnectionState {
	return &extint.ObjectConnectionState{
		ObjectId:   msg.ObjectID,
		FactoryId:  msg.FactoryID,
		ObjectType: extint.ObjectType(msg.ObjectType + 1),
		Connected:  msg.Connected,
	}
}
func CladObjectAvailableToProto(msg *gw_clad.ObjectAvailable) *extint.ObjectAvailable {
	return &extint.ObjectAvailable{
		FactoryId: msg.FactoryId,
	}
}
func CladObjectMovedToProto(msg *gw_clad.ObjectMoved) *extint.ObjectMoved {
	return &extint.ObjectMoved{
		Timestamp: msg.Timestamp,
		ObjectId:  msg.ObjectID,
	}
}
func CladObjectStoppedMovingToProto(msg *gw_clad.ObjectStoppedMoving) *extint.ObjectStoppedMoving {
	return &extint.ObjectStoppedMoving{
		Timestamp: msg.Timestamp,
		ObjectId:  msg.ObjectID,
	}
}
func CladObjectUpAxisChangedToProto(msg *gw_clad.ObjectUpAxisChanged) *extint.ObjectUpAxisChanged {
	// In clad, unknown is the final value
	// In proto, the convention is that 0 is unknown
	upAxis := extint.UpAxis_INVALID_AXIS
	if msg.UpAxis != gw_clad.UpAxis_UnknownAxis {
		upAxis = extint.UpAxis(msg.UpAxis + 1)
	}

	return &extint.ObjectUpAxisChanged{
		Timestamp: msg.Timestamp,
		ObjectId:  msg.ObjectID,
		UpAxis:    upAxis,
	}
}
func CladObjectTappedToProto(msg *gw_clad.ObjectTapped) *extint.ObjectTapped {
	return &extint.ObjectTapped{
		Timestamp: msg.Timestamp,
		ObjectId:  msg.ObjectID,
	}
}

func CladRobotObservedObjectToProto(msg *gw_clad.RobotObservedObject) *extint.RobotObservedObject {
	return &extint.RobotObservedObject{
		Timestamp:             msg.Timestamp,
		ObjectFamily:          extint.ObjectFamily(msg.ObjectFamily + 1),
		ObjectType:            extint.ObjectType(msg.ObjectType + 1),
		ObjectId:              msg.ObjectID,
		ImgRect:               CladCladRectToProto(&msg.ImgRect),
		Pose:                  CladPoseToProto(&msg.Pose),
		IsActive:              uint32(msg.IsActive),
		TopFaceOrientationRad: msg.TopFaceOrientationRad,
	}
}

func CladMemoryMapBeginToProtoNavMapInfo(msg *gw_clad.MemoryMapMessageBegin) *extint.NavMapInfo {
	return &extint.NavMapInfo{
		RootDepth:   int32(msg.RootDepth),
		RootSizeMm:  msg.RootSizeMm,
		RootCenterX: msg.RootCenterX,
		RootCenterY: msg.RootCenterY,
		RootCenterZ: 0.0,
	}
}

func CladMemoryMapQuadInfoToProto(msg *gw_clad.MemoryMapQuadInfo) *extint.NavMapQuadInfo {
	return &extint.NavMapQuadInfo{
		Content:   extint.NavNodeContentType(msg.Content), // Not incrementing this one because the CLAD enum has 0 as unknown
		Depth:     uint32(msg.Depth),
		ColorRgba: msg.ColorRGBA,
	}
}

func SendOnboardingComplete(in *extint.GatewayWrapper_OnboardingCompleteRequest) (*extint.OnboardingInputResponse, error) {
	f, responseChan := engineProtoManager.CreateChannel(&extint.GatewayWrapper_OnboardingCompleteResponse{}, 1)
	defer f()
	_, _, err := engineProtoManager.Write(&extint.GatewayWrapper{
		OneofMessageType: in,
	})
	if err != nil {
		return nil, err
	}
	completeResponse, ok := <-responseChan
	if !ok {
		return nil, grpc.Errorf(codes.Internal, "Failed to retrieve message")
	}
	return &extint.OnboardingInputResponse{
		Status: &extint.ResponseStatus{
			Code: extint.ResponseStatus_REQUEST_PROCESSING,
		},
		OneofMessageType: &extint.OnboardingInputResponse_OnboardingCompleteResponse{
			OnboardingCompleteResponse: completeResponse.GetOnboardingCompleteResponse(),
		},
	}, nil
}

func SendOnboardingWakeUpStarted(in *extint.GatewayWrapper_OnboardingWakeUpStartedRequest) (*extint.OnboardingInputResponse, error) {
	f, responseChan := engineProtoManager.CreateChannel(&extint.GatewayWrapper_OnboardingWakeUpStartedResponse{}, 1)
	defer f()
	_, _, err := engineProtoManager.Write(&extint.GatewayWrapper{
		OneofMessageType: in,
	})
	if err != nil {
		return nil, err
	}
	completeResponse, ok := <-responseChan
	if !ok {
		return nil, grpc.Errorf(codes.Internal, "Failed to retrieve message")
	}
	return &extint.OnboardingInputResponse{
		Status: &extint.ResponseStatus{
			Code: extint.ResponseStatus_REQUEST_PROCESSING,
		},
		OneofMessageType: &extint.OnboardingInputResponse_OnboardingWakeUpStartedResponse{
			OnboardingWakeUpStartedResponse: completeResponse.GetOnboardingWakeUpStartedResponse(),
		},
	}, nil
}

func SendOnboardingWakeUp(in *extint.GatewayWrapper_OnboardingWakeUpRequest) (*extint.OnboardingInputResponse, error) {
	f, responseChan := engineProtoManager.CreateChannel(&extint.GatewayWrapper_OnboardingWakeUpResponse{}, 1)
	defer f()
	_, _, err := engineProtoManager.Write(&extint.GatewayWrapper{
		OneofMessageType: in,
	})
	if err != nil {
		return nil, err
	}
	wakeUpResponse, ok := <-responseChan
	if !ok {
		return nil, grpc.Errorf(codes.Internal, "Failed to retrieve message")
	}
	return &extint.OnboardingInputResponse{
		Status: &extint.ResponseStatus{
			Code: extint.ResponseStatus_REQUEST_PROCESSING,
		},
		OneofMessageType: &extint.OnboardingInputResponse_OnboardingWakeUpResponse{
			OnboardingWakeUpResponse: wakeUpResponse.GetOnboardingWakeUpResponse(),
		},
	}, nil
}

func SendAppDisconnected() {
	msg := &extint.GatewayWrapper_AppDisconnected{
		AppDisconnected: &extint.AppDisconnected{},
	}
	engineProtoManager.Write(&extint.GatewayWrapper{
		OneofMessageType: msg,
	})
	// no error handling
}

func SendOnboardingSkipOnboarding(in *extint.GatewayWrapper_OnboardingSkipOnboarding) (*extint.OnboardingInputResponse, error) {
	_, _, err := engineProtoManager.Write(&extint.GatewayWrapper{
		OneofMessageType: in,
	})
	if err != nil {
		return nil, err
	}
	return &extint.OnboardingInputResponse{
		Status: &extint.ResponseStatus{
			Code: extint.ResponseStatus_REQUEST_PROCESSING,
		},
	}, nil
}

func SendOnboardingRestart(in *extint.GatewayWrapper_OnboardingRestart) (*extint.OnboardingInputResponse, error) {
	_, _, err := engineProtoManager.Write(&extint.GatewayWrapper{
		OneofMessageType: in,
	})
	if err != nil {
		return nil, err
	}
	return &extint.OnboardingInputResponse{
		Status: &extint.ResponseStatus{
			Code: extint.ResponseStatus_REQUEST_PROCESSING,
		},
	}, nil
}

func SendOnboardingSetPhase(in *extint.GatewayWrapper_OnboardingSetPhaseRequest) (*extint.OnboardingInputResponse, error) {
	f, responseChan := engineProtoManager.CreateChannel(&extint.GatewayWrapper_OnboardingSetPhaseResponse{}, 1)
	defer f()
	_, _, err := engineProtoManager.Write(&extint.GatewayWrapper{
		OneofMessageType: in,
	})
	if err != nil {
		return nil, err
	}
	completeResponse, ok := <-responseChan
	if !ok {
		return nil, grpc.Errorf(codes.Internal, "Failed to retrieve message")
	}
	return &extint.OnboardingInputResponse{
		Status: &extint.ResponseStatus{
			Code: extint.ResponseStatus_REQUEST_PROCESSING,
		},
		OneofMessageType: &extint.OnboardingInputResponse_OnboardingSetPhaseResponse{
			OnboardingSetPhaseResponse: completeResponse.GetOnboardingSetPhaseResponse(),
		},
	}, nil
}

func SendOnboardingPhaseProgress(in *extint.GatewayWrapper_OnboardingPhaseProgressRequest) (*extint.OnboardingInputResponse, error) {
	f, responseChan := engineProtoManager.CreateChannel(&extint.GatewayWrapper_OnboardingPhaseProgressResponse{}, 1)
	defer f()
	_, _, err := engineProtoManager.Write(&extint.GatewayWrapper{
		OneofMessageType: in,
	})
	if err != nil {
		return nil, err
	}
	completeResponse, ok := <-responseChan
	if !ok {
		return nil, grpc.Errorf(codes.Internal, "Failed to retrieve message")
	}
	return &extint.OnboardingInputResponse{
		Status: &extint.ResponseStatus{
			Code: extint.ResponseStatus_REQUEST_PROCESSING,
		},
		OneofMessageType: &extint.OnboardingInputResponse_OnboardingPhaseProgressResponse{
			OnboardingPhaseProgressResponse: completeResponse.GetOnboardingPhaseProgressResponse(),
		},
	}, nil
}

func SendOnboardingChargeInfo(in *extint.GatewayWrapper_OnboardingChargeInfoRequest) (*extint.OnboardingInputResponse, error) {
	f, responseChan := engineProtoManager.CreateChannel(&extint.GatewayWrapper_OnboardingChargeInfoResponse{}, 1)
	defer f()
	_, _, err := engineProtoManager.Write(&extint.GatewayWrapper{
		OneofMessageType: in,
	})
	if err != nil {
		return nil, err
	}
	completeResponse, ok := <-responseChan
	if !ok {
		return nil, grpc.Errorf(codes.Internal, "Failed to retrieve message")
	}
	return &extint.OnboardingInputResponse{
		Status: &extint.ResponseStatus{
			Code: extint.ResponseStatus_REQUEST_PROCESSING,
		},
		OneofMessageType: &extint.OnboardingInputResponse_OnboardingChargeInfoResponse{
			OnboardingChargeInfoResponse: completeResponse.GetOnboardingChargeInfoResponse(),
		},
	}, nil
}

func SendOnboardingMarkCompleteAndExit(in *extint.GatewayWrapper_OnboardingMarkCompleteAndExit) (*extint.OnboardingInputResponse, error) {
	_, _, err := engineProtoManager.Write(&extint.GatewayWrapper{
		OneofMessageType: in,
	})
	if err != nil {
		return nil, err
	}
	return &extint.OnboardingInputResponse{
		Status: &extint.ResponseStatus{
			Code: extint.ResponseStatus_REQUEST_PROCESSING,
		},
	}, nil
}

// The service definition.
// This must implement all the rpc functions defined in the external_interface proto file.
type rpcService struct{}

func (service *rpcService) ProtocolVersion(ctx context.Context, in *extint.ProtocolVersionRequest) (*extint.ProtocolVersionResponse, error) {
	response := &extint.ProtocolVersionResponse{
		HostVersion: int64(extint.ProtocolVersion_PROTOCOL_VERSION_CURRENT),
	}
	if in.ClientVersion < int64(extint.ProtocolVersion_PROTOCOL_VERSION_MINIMUM) {
		response.Result = extint.ProtocolVersionResponse_UNSUPPORTED
	} else {
		response.Result = extint.ProtocolVersionResponse_SUCCESS
	}
	return response, nil
}

func (service *rpcService) DriveWheels(ctx context.Context, in *extint.DriveWheelsRequest) (*extint.DriveWheelsResponse, error) {
	message := &extint.GatewayWrapper{
		OneofMessageType: &extint.GatewayWrapper_DriveWheelsRequest{
			DriveWheelsRequest: in,
		},
	}
	_, _, err := engineProtoManager.Write(message)
	if err != nil {
		return nil, err
	}
	return &extint.DriveWheelsResponse{
		Status: &extint.ResponseStatus{
			Code: extint.ResponseStatus_REQUEST_PROCESSING,
		},
	}, nil
}

// PlayAnimationTrigger intentionally waits for PlayAnimationResponse (not something like PlayAnimationTriggerResponse), because
// in the end, the engine is playing an animation.
func (service *rpcService) PlayAnimationTrigger(ctx context.Context, in *extint.PlayAnimationTriggerRequest) (*extint.PlayAnimationResponse, error) {
	f, animResponseChan := engineProtoManager.CreateChannel(&extint.GatewayWrapper_PlayAnimationResponse{}, 1)
	defer f()

	message := &extint.GatewayWrapper{
		OneofMessageType: &extint.GatewayWrapper_PlayAnimationTriggerRequest{
			PlayAnimationTriggerRequest: in,
		},
	}
	_, _, err := engineProtoManager.Write(message)
	if err != nil {
		return nil, err
	}

	setPlayAnimationResponse, ok := <-animResponseChan
	if !ok {
		return nil, grpc.Errorf(codes.Internal, "Failed to retrieve message")
	}
	response := setPlayAnimationResponse.GetPlayAnimationResponse()
	response.Status = &extint.ResponseStatus{
		Code: extint.ResponseStatus_RESPONSE_RECEIVED,
	}
	return response, nil
}

func (service *rpcService) PlayAnimation(ctx context.Context, in *extint.PlayAnimationRequest) (*extint.PlayAnimationResponse, error) {
	f, animResponseChan := engineProtoManager.CreateChannel(&extint.GatewayWrapper_PlayAnimationResponse{}, 1)
	defer f()

	message := &extint.GatewayWrapper{
		OneofMessageType: &extint.GatewayWrapper_PlayAnimationRequest{
			PlayAnimationRequest: in,
		},
	}
	_, _, err := engineProtoManager.Write(message)
	if err != nil {
		return nil, err
	}

	setPlayAnimationResponse, ok := <-animResponseChan
	if !ok {
		return nil, grpc.Errorf(codes.Internal, "Failed to retrieve message")
	}
	response := setPlayAnimationResponse.GetPlayAnimationResponse()
	response.Status = &extint.ResponseStatus{
		Code: extint.ResponseStatus_RESPONSE_RECEIVED,
	}
	return response, nil
}

func (service *rpcService) ListAnimations(ctx context.Context, in *extint.ListAnimationsRequest) (*extint.ListAnimationsResponse, error) {
	// 50 messages are sent per engine tick, however, in case it puts out multiple ticks before we drain, we need a buffer to hold lots o' data.
	delete_listener_callback, animationAvailableResponse := engineProtoManager.CreateChannel(&extint.GatewayWrapper_ListAnimationsResponse{}, 500)
	defer delete_listener_callback()

	message := &extint.GatewayWrapper{
		OneofMessageType: &extint.GatewayWrapper_ListAnimationsRequest{
			ListAnimationsRequest: in,
		},
	}

	_, _, err := engineProtoManager.Write(message)
	if err != nil {
		return nil, err
	}

	var anims []*extint.Animation

	done := false
	for done == false {
		select {
		case chanResponse, ok := <-animationAvailableResponse:
			if !ok {
				return nil, grpc.Errorf(codes.Internal, "Failed to retrieve message")
			}
			for _, anim := range chanResponse.GetListAnimationsResponse().AnimationNames {
				animName := anim.GetName()
				// Don't change endOfAnimationList - it's what we'll receive from the .cpp sender.
				if animName == endOfAnimationList {
					done = true
				} else {
					if strings.Contains(animName, "_avs_") {
						// VIC-11583 All Alexa animation names contain "_avs_". Prevent these animations from reaching the SDK.
						continue
					}
					var newAnim = extint.Animation{
						Name: animName,
					}
					anims = append(anims, &newAnim)
				}
			}
		case <-time.After(5 * time.Second):
			return nil, grpc.Errorf(codes.DeadlineExceeded, "ListAnimations request timed out")
		}
	}

	return &extint.ListAnimationsResponse{
		Status: &extint.ResponseStatus{
			Code: extint.ResponseStatus_RESPONSE_RECEIVED,
		},
		AnimationNames: anims,
	}, nil
}

func (service *rpcService) ListAnimationTriggers(ctx context.Context, in *extint.ListAnimationTriggersRequest) (*extint.ListAnimationTriggersResponse, error) {
	delete_listener_callback, animationTriggerAvailableResponse := engineProtoManager.CreateChannel(&extint.GatewayWrapper_ListAnimationTriggersResponse{}, 500)
	defer delete_listener_callback()

	message := &extint.GatewayWrapper{
		OneofMessageType: &extint.GatewayWrapper_ListAnimationTriggersRequest{
			ListAnimationTriggersRequest: in,
		},
	}

	_, _, err := engineProtoManager.Write(message)
	if err != nil {
		return nil, err
	}

	var animTriggers []*extint.AnimationTrigger

	done := false
	for done == false {
		select {
		case chanResponse, ok := <-animationTriggerAvailableResponse:
			if !ok {
				return nil, grpc.Errorf(codes.Internal, "Failed to retrieve message")
			}
			for _, animTrigger := range chanResponse.GetListAnimationTriggersResponse().AnimationTriggerNames {
				animTriggerName := animTrigger.GetName()
				// Don't change endOfAnimationList - it's what we'll receive from the .cpp sender.
				if animTriggerName == endOfAnimationList {
					done = true
				} else {
					animTriggerNameLower := strings.ToLower(animTriggerName)
					if strings.Contains(animTriggerNameLower, "deprecated") || strings.Contains(animTriggerNameLower, "alexa") {
						// Prevent animation triggers that are deprecated or are for alexa from reaching the SDK
						continue
					}
					newAnimTrigger := extint.AnimationTrigger{
						Name: animTriggerName,
					}
					animTriggers = append(animTriggers, &newAnimTrigger)
				}
			}
		case <-time.After(10 * time.Second):
			return nil, grpc.Errorf(codes.DeadlineExceeded, "ListAnimationTriggers request timed out")
		}
	}

	return &extint.ListAnimationTriggersResponse{
		Status: &extint.ResponseStatus{
			Code: extint.ResponseStatus_RESPONSE_RECEIVED,
		},
		AnimationTriggerNames: animTriggers,
	}, nil
}

func (service *rpcService) MoveHead(ctx context.Context, in *extint.MoveHeadRequest) (*extint.MoveHeadResponse, error) {
	_, err := engineCladManager.Write(ProtoMoveHeadToClad(in))
	if err != nil {
		return nil, err
	}
	return &extint.MoveHeadResponse{
		Status: &extint.ResponseStatus{
			Code: extint.ResponseStatus_REQUEST_PROCESSING,
		},
	}, nil
}

func (service *rpcService) MoveLift(ctx context.Context, in *extint.MoveLiftRequest) (*extint.MoveLiftResponse, error) {
	_, err := engineCladManager.Write(ProtoMoveLiftToClad(in))
	if err != nil {
		return nil, err
	}
	return &extint.MoveLiftResponse{
		Status: &extint.ResponseStatus{
			Code: extint.ResponseStatus_REQUEST_PROCESSING,
		},
	}, nil
}

func (service *rpcService) StopAllMotors(ctx context.Context, in *extint.StopAllMotorsRequest) (*extint.StopAllMotorsResponse, error) {
	message := &extint.GatewayWrapper{
		OneofMessageType: &extint.GatewayWrapper_StopAllMotorsRequest{
			StopAllMotorsRequest: in,
		},
	}
	_, _, err := engineProtoManager.Write(message)
	if err != nil {
		return nil, err
	}
	return &extint.StopAllMotorsResponse{
		Status: &extint.ResponseStatus{
			Code: extint.ResponseStatus_REQUEST_PROCESSING,
		},
	}, nil
}

func (service *rpcService) CancelBehavior(ctx context.Context, in *extint.CancelBehaviorRequest) (*extint.CancelBehaviorResponse, error) {
	message := &extint.GatewayWrapper{
		OneofMessageType: &extint.GatewayWrapper_CancelBehaviorRequest{
			CancelBehaviorRequest: in,
		},
	}
	_, _, err := engineProtoManager.Write(message)
	if err != nil {
		return nil, err
	}
	return &extint.CancelBehaviorResponse{
		Status: &extint.ResponseStatus{
			Code: extint.ResponseStatus_REQUEST_PROCESSING,
		},
	}, nil
}

func (service *rpcService) CancelActionByIdTag(ctx context.Context, in *extint.CancelActionByIdTagRequest) (*extint.CancelActionByIdTagResponse, error) {
	message := &extint.GatewayWrapper{
		OneofMessageType: &extint.GatewayWrapper_CancelActionByIdTagRequest{
			CancelActionByIdTagRequest: in,
		},
	}
	_, _, err := engineProtoManager.Write(message)
	if err != nil {
		return nil, err
	}
	return &extint.CancelActionByIdTagResponse{
		Status: &extint.ResponseStatus{
			Code: extint.ResponseStatus_REQUEST_PROCESSING,
		},
	}, nil
}

func SendFaceDataAsChunks(in *extint.DisplayFaceImageRGBRequest, chunkCount int, pixelsPerChunk int, totalPixels int) error {
	var convertedUint16Data [faceImagePixelsPerChunk]uint16

	// cycle until we run out of bytes to transfer
	for i := 0; i < chunkCount; i++ {
		pixelCount := faceImagePixelsPerChunk
		if i == chunkCount-1 {
			pixelCount = totalPixels - faceImagePixelsPerChunk*i
		}

		firstByte := (pixelsPerChunk * 2) * i
		finalByte := firstByte + (pixelCount * 2)
		slicedBinaryData := in.FaceData[firstByte:finalByte] // TODO: Make this not implode on empty

		for j := 0; j < pixelCount; j++ {
			uintAsBytes := slicedBinaryData[j*2 : j*2+2]
			convertedUint16Data[j] = binary.BigEndian.Uint16(uintAsBytes)
		}

		// Copy a subset of the pixels to the bytes?
		message := FaceImageChunkToClad(convertedUint16Data, uint16(pixelCount), uint8(i), uint8(chunkCount), in.DurationMs, in.InterruptRunning)

		_, err := engineCladManager.Write(message)
		if err != nil {
			return err
		}
	}

	return nil
}

func (service *rpcService) DisplayFaceImageRGB(ctx context.Context, in *extint.DisplayFaceImageRGBRequest) (*extint.DisplayFaceImageRGBResponse, error) {
	const totalPixels = 17664
	chunkCount := (totalPixels + faceImagePixelsPerChunk + 1) / faceImagePixelsPerChunk

	SendFaceDataAsChunks(in, chunkCount, faceImagePixelsPerChunk, totalPixels)

	return &extint.DisplayFaceImageRGBResponse{
		Status: &extint.ResponseStatus{
			Code: extint.ResponseStatus_REQUEST_PROCESSING,
		},
	}, nil
}

func (service *rpcService) AppIntent(ctx context.Context, in *extint.AppIntentRequest) (*extint.AppIntentResponse, error) {
	_, err := engineCladManager.Write(ProtoAppIntentToClad(in))
	if err != nil {
		return nil, err
	}
	return &extint.AppIntentResponse{
		Status: &extint.ResponseStatus{
			Code: extint.ResponseStatus_REQUEST_PROCESSING,
		},
	}, nil
}

func (service *rpcService) CancelFaceEnrollment(ctx context.Context, in *extint.CancelFaceEnrollmentRequest) (*extint.CancelFaceEnrollmentResponse, error) {
	_, err := engineCladManager.Write(ProtoCancelFaceEnrollmentToClad(in))
	if err != nil {
		return nil, err
	}
	return &extint.CancelFaceEnrollmentResponse{
		Status: &extint.ResponseStatus{
			Code: extint.ResponseStatus_REQUEST_PROCESSING,
		},
	}, nil
}

func (service *rpcService) RequestEnrolledNames(ctx context.Context, in *extint.RequestEnrolledNamesRequest) (*extint.RequestEnrolledNamesResponse, error) {
	f, enrolledNamesResponse := engineCladManager.CreateChannel(gw_clad.MessageRobotToExternalTag_EnrolledNamesResponse, 1)
	defer f()

	_, err := engineCladManager.Write(ProtoRequestEnrolledNamesToClad(in))
	if err != nil {
		return nil, err
	}
	names, ok := <-enrolledNamesResponse
	if !ok {
		return nil, grpc.Errorf(codes.Internal, "Failed to retrieve message")
	}
	var faces []*extint.LoadedKnownFace
	for _, element := range names.GetEnrolledNamesResponse().Faces {
		var newFace = extint.LoadedKnownFace{
			SecondsSinceFirstEnrolled: element.SecondsSinceFirstEnrolled,
			SecondsSinceLastUpdated:   element.SecondsSinceLastUpdated,
			SecondsSinceLastSeen:      element.SecondsSinceLastSeen,
			LastSeenSecondsSinceEpoch: element.LastSeenSecondsSinceEpoch,
			FaceId:                    element.FaceID,
			Name:                      element.Name,
		}
		faces = append(faces, &newFace)
	}
	return &extint.RequestEnrolledNamesResponse{
		Status: &extint.ResponseStatus{
			Code: extint.ResponseStatus_RESPONSE_RECEIVED,
		},
		Faces: faces,
	}, nil
}

// TODO Wait for response RobotRenamedEnrolledFace
func (service *rpcService) UpdateEnrolledFaceByID(ctx context.Context, in *extint.UpdateEnrolledFaceByIDRequest) (*extint.UpdateEnrolledFaceByIDResponse, error) {
	_, err := engineCladManager.Write(ProtoUpdateEnrolledFaceByIDToClad(in))
	if err != nil {
		return nil, err
	}
	return &extint.UpdateEnrolledFaceByIDResponse{
		Status: &extint.ResponseStatus{
			Code: extint.ResponseStatus_REQUEST_PROCESSING,
		},
	}, nil
}

// TODO Wait for response RobotRenamedEnrolledFace
func (service *rpcService) EraseEnrolledFaceByID(ctx context.Context, in *extint.EraseEnrolledFaceByIDRequest) (*extint.EraseEnrolledFaceByIDResponse, error) {
	_, err := engineCladManager.Write(ProtoEraseEnrolledFaceByIDToClad(in))
	if err != nil {
		return nil, err
	}
	return &extint.EraseEnrolledFaceByIDResponse{
		Status: &extint.ResponseStatus{
			Code: extint.ResponseStatus_REQUEST_PROCESSING,
		},
	}, nil
}

// TODO Wait for response RobotErasedAllEnrolledFaces
func (service *rpcService) EraseAllEnrolledFaces(ctx context.Context, in *extint.EraseAllEnrolledFacesRequest) (*extint.EraseAllEnrolledFacesResponse, error) {
	_, err := engineCladManager.Write(ProtoEraseAllEnrolledFacesToClad(in))
	if err != nil {
		return nil, err
	}
	return &extint.EraseAllEnrolledFacesResponse{
		Status: &extint.ResponseStatus{
			Code: extint.ResponseStatus_REQUEST_PROCESSING,
		},
	}, nil
}

func (service *rpcService) SetFaceToEnroll(ctx context.Context, in *extint.SetFaceToEnrollRequest) (*extint.SetFaceToEnrollResponse, error) {
	f, responseChan := engineProtoManager.CreateChannel(&extint.GatewayWrapper_SetFaceToEnrollResponse{}, 1)
	defer f()

	_, _, err := engineProtoManager.Write(&extint.GatewayWrapper{
		OneofMessageType: &extint.GatewayWrapper_SetFaceToEnrollRequest{
			SetFaceToEnrollRequest: in,
		},
	})
	if err != nil {
		return nil, err
	}
	setFaceToEnrollResponse, ok := <-responseChan
	if !ok {
		return nil, grpc.Errorf(codes.Internal, "Failed to retrieve message")
	}
	response := setFaceToEnrollResponse.GetSetFaceToEnrollResponse()
	response.Status = &extint.ResponseStatus{
		Code: extint.ResponseStatus_RESPONSE_RECEIVED,
	}
	return response, nil
}

func (service *rpcService) EnrollFace(ctx context.Context, in *extint.EnrollFaceRequest) (*extint.EnrollFaceResponse, error) {
	f, responseChan := engineProtoManager.CreateChannel(&extint.GatewayWrapper_EnrollFaceResponse{}, 1)
	defer f()

	_, _, err := engineProtoManager.Write(&extint.GatewayWrapper{
		OneofMessageType: &extint.GatewayWrapper_EnrollFaceRequest{
			EnrollFaceRequest: in,
		},
	})
	if err != nil {
		return nil, err
	}
	enrollFaceResponse, ok := <-responseChan
	if !ok {
		return nil, grpc.Errorf(codes.Internal, "Failed to retrieve message")
	}
	response := enrollFaceResponse.GetEnrollFaceResponse()
	response.Status = &extint.ResponseStatus{
		Code: extint.ResponseStatus_RESPONSE_RECEIVED,
	}
	return response, nil
}

func isMember(needle string, haystack []string) bool {
	for _, word := range haystack {
		if word == needle {
			return true
		}
	}
	return false
}

func checkFilters(event *extint.Event, whiteList, blackList *extint.FilterList) bool {
	if whiteList == nil && blackList == nil {
		return true
	}
	props := &proto.Properties{}
	val := reflect.ValueOf(event.GetEventType())
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}
	props.Parse(val.Type().Field(0).Tag.Get("protobuf"))
	responseType := props.OrigName

	if whiteList != nil && isMember(responseType, whiteList.List) {
		return true
	}
	if blackList != nil && !isMember(responseType, blackList.List) {
		return true
	}
	return false
}

// Should be called on WiFi connect.
func (service *rpcService) onConnect(id string) {
	// Call DAS WiFi connection event to indicate start of a WiFi connection.
	// Log the connection id for the primary connection, which is the first person to connect.
	log.Das("wifi_conn_id.start", (&log.DasFields{}).SetStrings(id))
}

// Should be called on WiFi disconnect.
func (service *rpcService) onDisconnect() {
	// Message engine that app disconnected
	SendAppDisconnected()
	// Call DAS WiFi connection event to indicate stop of a WiFi connection
	log.Das("wifi_conn_id.stop", (&log.DasFields{}).SetStrings(""))
	connectionId = ""
}

func (service *rpcService) checkConnectionID(id string) bool {
	connectionIdLock.Lock()
	defer connectionIdLock.Unlock()
	if len(connectionId) != 0 && id != connectionId {
		log.Println("Connection id already set: current='%s', incoming='%s'", connectionId, id)
		return false
	}
	// Check whether we are in Webots.
	if IsOnRobot {
		f, responseChan := switchboardManager.CreateChannel(gw_clad.SwitchboardResponseTag_ExternalConnectionResponse, 1)
		defer f()
		switchboardManager.Write(gw_clad.NewSwitchboardRequestWithExternalConnectionRequest(&gw_clad.ExternalConnectionRequest{}))

		response, ok := <-responseChan
		if !ok {
			log.Println("Failed to receive ConnectionID response from vic-switchboard")
			return false
		}
		connectionResponse := response.GetExternalConnectionResponse()

		// IsConnected shows whether switchboard is connected over ble.
		// Detect if someone else is connected.
		if connectionResponse.IsConnected && connectionResponse.ConnectionId != id {
			// Someone is connected over BLE and they are not the primary connection.
			// We return false so the app can tell you not to connect.
			log.Printf("Detected mismatched BLE connection id: BLE='%s', incoming='%s'\n", connectionResponse.ConnectionId, id)
			return false
		}
	}
	connectionId = id
	return true
}

// SDK-only message to pass version info for device OS, Python version, etc.
func (service *rpcService) SDKInitialization(ctx context.Context, in *extint.SDKInitializationRequest) (*extint.SDKInitializationResponse, error) {
	log.Das("sdk.module_version", (&log.DasFields{}).SetStrings(in.SdkModuleVersion))
	log.Das("sdk.python_version", (&log.DasFields{}).SetStrings(in.PythonVersion))
	log.Das("sdk.python_implementation", (&log.DasFields{}).SetStrings(in.PythonImplementation))
	log.Das("sdk.os_version", (&log.DasFields{}).SetStrings(in.OsVersion))
	log.Das("sdk.cpu_version", (&log.DasFields{}).SetStrings(in.CpuVersion))

	return &extint.SDKInitializationResponse{
		Status: &extint.ResponseStatus{
			Code: extint.ResponseStatus_REQUEST_PROCESSING,
		},
	}, nil
}

// Long running message for sending events to listening sdk users
func (service *rpcService) EventStream(in *extint.EventRequest, stream extint.ExternalInterface_EventStreamServer) error {
	isPrimary := service.checkConnectionID(in.ConnectionId)
	if isPrimary {
		service.onConnect(connectionId)
		defer service.onDisconnect()
	}
	resp := &extint.EventResponse{
		Status: &extint.ResponseStatus{
			Code: extint.ResponseStatus_RESPONSE_RECEIVED,
		},
		Event: &extint.Event{
			EventType: &extint.Event_ConnectionResponse{
				ConnectionResponse: &extint.ConnectionResponse{
					IsPrimary: isPrimary,
				},
			},
		},
	}
	err := stream.Send(resp)
	if err != nil {
		log.Println("Closing Event stream (on send):", err)
		return err
	} else if err = stream.Context().Err(); err != nil {
		log.Println("Closing Event stream:", err)
		// This is the case where the user disconnects the stream
		// We should still return the err in case the user doesn't think they disconnected
		return err
	}

	if isPrimary {
		log.Printf("EventStream: Sent primary connection response '%s'\n", connectionId)
	} else {
		log.Printf("EventStream: Sent secondary connection response given='%s', current='%s'\n", in.ConnectionId, connectionId)
	}

	f, eventsChannel := engineProtoManager.CreateChannel(&extint.GatewayWrapper_Event{}, 512)
	defer f()

	ping := extint.EventResponse{
		Event: &extint.Event{
			EventType: &extint.Event_KeepAlive{
				KeepAlive: &extint.KeepAlivePing{},
			},
		},
	}

	whiteList := in.GetWhiteList()
	blackList := in.GetBlackList()

	pingTicker := time.Tick(time.Second)

	for {
		select {
		case response, ok := <-eventsChannel:
			if !ok {
				return grpc.Errorf(codes.Internal, "EventStream: event channel closed")
			}
			event := response.GetEvent()
			if checkFilters(event, whiteList, blackList) {
				if logVerbose {
					log.Printf("EventStream: Sending event to client: %#v\n", *event)
				}
				eventResponse := &extint.EventResponse{
					Event: event,
				}
				if err := stream.Send(eventResponse); err != nil {
					log.Println("Closing Event stream (on send):", err)
					return err
				} else if err = stream.Context().Err(); err != nil {
					log.Println("Closing Event stream:", err)
					// This is the case where the user disconnects the stream
					// We should still return the err in case the user doesn't think they disconnected
					return err
				}
			}
		case <-pingTicker: // ping to check connection liveness after one second.
			if err := stream.Send(&ping); err != nil {
				log.Println("Closing Event stream (on send):", err)
				return err
			} else if err = stream.Context().Err(); err != nil {
				log.Println("Closing Event stream:", err)
				// This is the case where the user disconnects the stream
				// We should still return the err in case the user doesn't think they disconnected
				return err
			}
		}
	}
	return nil
}

func (service *rpcService) BehaviorRequestToGatewayWrapper(request *extint.BehaviorControlRequest) (*extint.GatewayWrapper, error) {
	msg := &extint.GatewayWrapper{}

	switch x := request.RequestType.(type) {
	case *extint.BehaviorControlRequest_ControlRelease:
		msg.OneofMessageType = &extint.GatewayWrapper_ControlRelease{
			ControlRelease: request.GetControlRelease(),
		}
	case *extint.BehaviorControlRequest_ControlRequest:
		msg.OneofMessageType = &extint.GatewayWrapper_ControlRequest{
			ControlRequest: request.GetControlRequest(),
		}
	default:
		return nil, grpc.Errorf(codes.InvalidArgument, "BehaviorControlRequest.ControlRequest has unexpected type %T", x)
	}
	return msg, nil
}

func (service *rpcService) BehaviorControlRequestHandler(in extint.ExternalInterface_BehaviorControlServer, done chan struct{}, connID uint64) {
	defer close(done)
	defer engineProtoManager.WriteWithID(&extint.GatewayWrapper{
		OneofMessageType: &extint.GatewayWrapper_ControlRelease{
			ControlRelease: &extint.ControlRelease{},
		},
	}, connID)

	for {
		request, err := in.Recv()
		if err != nil {
			log.Printf("BehaviorControlRequestHandler.close: %s\n", err.Error())
			return
		}
		log.Println("BehaviorControl Incoming Request:", request)

		msg, err := service.BehaviorRequestToGatewayWrapper(request)
		if err != nil {
			log.Println(err)
			return
		}
		_, err = engineProtoManager.WriteWithID(msg, connID)
		if err != nil {
			return
		}
	}
}

func (service *rpcService) BehaviorControlResponseHandler(out extint.ExternalInterface_AssumeBehaviorControlServer, responses chan extint.GatewayWrapper, done chan struct{}, connID uint64) error {
	ping := extint.BehaviorControlResponse{
		ResponseType: &extint.BehaviorControlResponse_KeepAlive{
			KeepAlive: &extint.KeepAlivePing{},
		},
	}

	pingTicker := time.Tick(time.Second)

	for {
		select {
		case <-done:
			return nil
		case response, ok := <-responses:
			if !ok {
				return grpc.Errorf(codes.Internal, "Failed to retrieve message")
			}
			if connID != response.GetConnectionId() {
				continue
			}
			msg := response.GetBehaviorControlResponse()
			log.Println("BehaviorControl Incoming BehaviorControlResponse message:", msg)
			if err := out.Send(msg); err != nil {
				log.Printf("Closing BehaviorControl stream (on send): %s connID %d\n", err.Error(), connID)
				return err
			} else if err = out.Context().Err(); err != nil {
				// This is the case where the user disconnects the stream
				// We should still return the err in case the user doesn't think they disconnected
				log.Printf("Closing BehaviorControl stream: %s connID %d\n", err, connID)
				return err
			}
		case <-pingTicker: // ping to check connection liveness after one second.
			if err := out.Send(&ping); err != nil {
				log.Printf("Closing BehaviorControl stream (on send): %s connID %d\n", err.Error(), connID)
				return err
			} else if err = out.Context().Err(); err != nil {
				// This is the case where the user disconnects the stream
				// We should still return the err in case the user doesn't think they disconnected
				log.Printf("Closing BehaviorControl stream: %s connID %d\n", err, connID)
				return err
			}
		}
	}
	return nil
}

// SDK-only method. SDK DAS connect/disconnect events are sent from here.
func (service *rpcService) BehaviorControl(bidirectionalStream extint.ExternalInterface_BehaviorControlServer) error {
	sdkStartTime := time.Now()

	numCommandsSentFromSDK = 0

	log.Das("sdk.connection_started", (&log.DasFields{}).SetStrings(""))

	defer func() {
		sdkElapsedSeconds := time.Since(sdkStartTime)
		log.Das("sdk.connection_ended", (&log.DasFields{}).SetStrings(sdkElapsedSeconds.String(), fmt.Sprint(numCommandsSentFromSDK)))
		numCommandsSentFromSDK = 0
	}()

	done := make(chan struct{})
	connID := rand.Uint64()

	f, behaviorStatus := engineProtoManager.CreateChannel(&extint.GatewayWrapper_BehaviorControlResponse{}, 1)
	defer f()

	go service.BehaviorControlRequestHandler(bidirectionalStream, done, connID)
	return service.BehaviorControlResponseHandler(bidirectionalStream, behaviorStatus, done, connID)
}

func (service *rpcService) AssumeBehaviorControl(in *extint.BehaviorControlRequest, out extint.ExternalInterface_AssumeBehaviorControlServer) error {
	done := make(chan struct{})

	f, behaviorStatus := engineProtoManager.CreateChannel(&extint.GatewayWrapper_BehaviorControlResponse{}, 1)
	defer f()

	msg, err := service.BehaviorRequestToGatewayWrapper(in)
	if err != nil {
		return err
	}

	_, connID, err := engineProtoManager.Write(msg)
	if err != nil {
		return err
	}

	return service.BehaviorControlResponseHandler(out, behaviorStatus, done, connID)
}

func (service *rpcService) DriveOffCharger(ctx context.Context, in *extint.DriveOffChargerRequest) (*extint.DriveOffChargerResponse, error) {
	f, responseChan := engineProtoManager.CreateChannel(&extint.GatewayWrapper_DriveOffChargerResponse{}, 1)
	defer f()

	_, _, err := engineProtoManager.Write(&extint.GatewayWrapper{
		OneofMessageType: &extint.GatewayWrapper_DriveOffChargerRequest{
			DriveOffChargerRequest: in,
		},
	})
	if err != nil {
		return nil, err
	}
	driveOffChargerResponse, ok := <-responseChan
	if !ok {
		return nil, grpc.Errorf(codes.Internal, "Failed to retrieve message")
	}
	response := driveOffChargerResponse.GetDriveOffChargerResponse()
	response.Status = &extint.ResponseStatus{
		Code: extint.ResponseStatus_RESPONSE_RECEIVED,
	}
	return response, nil
}

func (service *rpcService) DriveOnCharger(ctx context.Context, in *extint.DriveOnChargerRequest) (*extint.DriveOnChargerResponse, error) {
	f, responseChan := engineProtoManager.CreateChannel(&extint.GatewayWrapper_DriveOnChargerResponse{}, 1)
	defer f()

	_, _, err := engineProtoManager.Write(&extint.GatewayWrapper{
		OneofMessageType: &extint.GatewayWrapper_DriveOnChargerRequest{
			DriveOnChargerRequest: in,
		},
	})
	if err != nil {
		return nil, err
	}
	driveOnChargerResponse, ok := <-responseChan
	if !ok {
		return nil, grpc.Errorf(codes.Internal, "Failed to retrieve message")
	}
	response := driveOnChargerResponse.GetDriveOnChargerResponse()
	response.Status = &extint.ResponseStatus{
		Code: extint.ResponseStatus_RESPONSE_RECEIVED,
	}
	return response, nil
}

func (service *rpcService) FindFaces(ctx context.Context, in *extint.FindFacesRequest) (*extint.FindFacesResponse, error) {
	f, responseChan := engineProtoManager.CreateChannel(&extint.GatewayWrapper_FindFacesResponse{}, 1)
	defer f()

	_, _, err := engineProtoManager.Write(&extint.GatewayWrapper{
		OneofMessageType: &extint.GatewayWrapper_FindFacesRequest{
			FindFacesRequest: in,
		},
	})
	if err != nil {
		return nil, err
	}
	findFacesResponse, ok := <-responseChan
	if !ok {
		return nil, grpc.Errorf(codes.Internal, "Failed to retrieve message")
	}
	response := findFacesResponse.GetFindFacesResponse()
	response.Status = &extint.ResponseStatus{
		Code: extint.ResponseStatus_RESPONSE_RECEIVED,
	}
	return response, nil
}

func (service *rpcService) LookAroundInPlace(ctx context.Context, in *extint.LookAroundInPlaceRequest) (*extint.LookAroundInPlaceResponse, error) {
	f, responseChan := engineProtoManager.CreateChannel(&extint.GatewayWrapper_LookAroundInPlaceResponse{}, 1)
	defer f()

	_, _, err := engineProtoManager.Write(&extint.GatewayWrapper{
		OneofMessageType: &extint.GatewayWrapper_LookAroundInPlaceRequest{
			LookAroundInPlaceRequest: in,
		},
	})
	if err != nil {
		return nil, err
	}
	lookAroundInPlaceResponse, ok := <-responseChan
	if !ok {
		return nil, grpc.Errorf(codes.Internal, "Failed to retrieve message")
	}
	response := lookAroundInPlaceResponse.GetLookAroundInPlaceResponse()
	response.Status = &extint.ResponseStatus{
		Code: extint.ResponseStatus_RESPONSE_RECEIVED,
	}
	return response, nil
}

func (service *rpcService) RollBlock(ctx context.Context, in *extint.RollBlockRequest) (*extint.RollBlockResponse, error) {
	f, responseChan := engineProtoManager.CreateChannel(&extint.GatewayWrapper_RollBlockResponse{}, 1)
	defer f()

	_, _, err := engineProtoManager.Write(&extint.GatewayWrapper{
		OneofMessageType: &extint.GatewayWrapper_RollBlockRequest{
			RollBlockRequest: in,
		},
	})
	if err != nil {
		return nil, err
	}
	rollBlockResponse, ok := <-responseChan
	if !ok {
		return nil, grpc.Errorf(codes.Internal, "Failed to retrieve message")
	}
	response := rollBlockResponse.GetRollBlockResponse()
	response.Status = &extint.ResponseStatus{
		Code: extint.ResponseStatus_RESPONSE_RECEIVED,
	}
	return response, nil
}

// Request the current robot onboarding status
func (service *rpcService) GetOnboardingState(ctx context.Context, in *extint.OnboardingStateRequest) (*extint.OnboardingStateResponse, error) {
	f, responseChan := engineProtoManager.CreateChannel(&extint.GatewayWrapper_OnboardingState{}, 1)
	defer f()
	_, _, err := engineProtoManager.Write(&extint.GatewayWrapper{
		OneofMessageType: &extint.GatewayWrapper_OnboardingStateRequest{
			OnboardingStateRequest: in,
		},
	})
	if err != nil {
		return nil, err
	}
	onboardingState, ok := <-responseChan
	if !ok {
		return nil, grpc.Errorf(codes.Internal, "Failed to retrieve message")
	}
	return &extint.OnboardingStateResponse{
		Status: &extint.ResponseStatus{
			Code: extint.ResponseStatus_RESPONSE_RECEIVED,
		},
		OnboardingState: onboardingState.GetOnboardingState(),
	}, nil
}

func (service *rpcService) SendOnboardingInput(ctx context.Context, in *extint.OnboardingInputRequest) (*extint.OnboardingInputResponse, error) {
	// oneof_message_type
	switch x := in.OneofMessageType.(type) {
	case *extint.OnboardingInputRequest_OnboardingCompleteRequest:
		return SendOnboardingComplete(&extint.GatewayWrapper_OnboardingCompleteRequest{
			OnboardingCompleteRequest: in.GetOnboardingCompleteRequest(),
		})
	case *extint.OnboardingInputRequest_OnboardingWakeUpStartedRequest:
		return SendOnboardingWakeUpStarted(&extint.GatewayWrapper_OnboardingWakeUpStartedRequest{
			OnboardingWakeUpStartedRequest: in.GetOnboardingWakeUpStartedRequest(),
		})
	case *extint.OnboardingInputRequest_OnboardingWakeUpRequest:
		return SendOnboardingWakeUp(&extint.GatewayWrapper_OnboardingWakeUpRequest{
			OnboardingWakeUpRequest: in.GetOnboardingWakeUpRequest(),
		})
	case *extint.OnboardingInputRequest_OnboardingSkipOnboarding:
		return SendOnboardingSkipOnboarding(&extint.GatewayWrapper_OnboardingSkipOnboarding{
			OnboardingSkipOnboarding: in.GetOnboardingSkipOnboarding(),
		})
	case *extint.OnboardingInputRequest_OnboardingRestart:
		return SendOnboardingRestart(&extint.GatewayWrapper_OnboardingRestart{
			OnboardingRestart: in.GetOnboardingRestart(),
		})
	case *extint.OnboardingInputRequest_OnboardingSetPhaseRequest:
		return SendOnboardingSetPhase(&extint.GatewayWrapper_OnboardingSetPhaseRequest{
			OnboardingSetPhaseRequest: in.GetOnboardingSetPhaseRequest(),
		})
	case *extint.OnboardingInputRequest_OnboardingPhaseProgressRequest:
		return SendOnboardingPhaseProgress(&extint.GatewayWrapper_OnboardingPhaseProgressRequest{
			OnboardingPhaseProgressRequest: in.GetOnboardingPhaseProgressRequest(),
		})
	case *extint.OnboardingInputRequest_OnboardingChargeInfoRequest:
		return SendOnboardingChargeInfo(&extint.GatewayWrapper_OnboardingChargeInfoRequest{
			OnboardingChargeInfoRequest: in.GetOnboardingChargeInfoRequest(),
		})
	case *extint.OnboardingInputRequest_OnboardingMarkCompleteAndExit:
		return SendOnboardingMarkCompleteAndExit(&extint.GatewayWrapper_OnboardingMarkCompleteAndExit{
			OnboardingMarkCompleteAndExit: in.GetOnboardingMarkCompleteAndExit(),
		})
	default:
		return nil, grpc.Errorf(codes.InvalidArgument, "OnboardingInputRequest.OneofMessageType has unexpected type %T", x)
	}
}

func (service *rpcService) PhotosInfo(ctx context.Context, in *extint.PhotosInfoRequest) (*extint.PhotosInfoResponse, error) {
	f, responseChan := engineProtoManager.CreateChannel(&extint.GatewayWrapper_PhotosInfoResponse{}, 1)
	defer f()

	_, _, err := engineProtoManager.Write(&extint.GatewayWrapper{
		OneofMessageType: &extint.GatewayWrapper_PhotosInfoRequest{
			PhotosInfoRequest: in,
		},
	})
	if err != nil {
		return nil, err
	}
	payload, ok := <-responseChan
	if !ok {
		return nil, grpc.Errorf(codes.Internal, "Failed to retrieve message")
	}
	infoResponse := payload.GetPhotosInfoResponse()
	infoResponse.Status = &extint.ResponseStatus{
		Code: extint.ResponseStatus_RESPONSE_RECEIVED,
	}
	return infoResponse, nil
}

func SendImageHelper(fullpath string) ([]byte, error) {
	log.Println("Reading file at", fullpath)
	dat, err := ioutil.ReadFile(fullpath)
	if err != nil {
		log.Println("Error reading file ", fullpath)
		return nil, err
	}
	return dat, nil
}

func (service *rpcService) Photo(ctx context.Context, in *extint.PhotoRequest) (*extint.PhotoResponse, error) {
	f, responseChan := engineProtoManager.CreateChannel(&extint.GatewayWrapper_PhotoPathMessage{}, 1)
	defer f()

	_, _, err := engineProtoManager.Write(&extint.GatewayWrapper{
		OneofMessageType: &extint.GatewayWrapper_PhotoRequest{
			PhotoRequest: in,
		},
	})
	if err != nil {
		return nil, err
	}
	payload, ok := <-responseChan
	if !ok {
		return nil, grpc.Errorf(codes.Internal, "Failed to retrieve message")
	}
	if !payload.GetPhotoPathMessage().GetSuccess() {
		return &extint.PhotoResponse{
			Status: &extint.ResponseStatus{
				Code: extint.ResponseStatus_NOT_FOUND,
			},
			Success: false,
		}, err
	}
	imageData, err := SendImageHelper(payload.GetPhotoPathMessage().GetFullPath())
	if err != nil {
		return &extint.PhotoResponse{
			Status: &extint.ResponseStatus{
				Code: extint.ResponseStatus_NOT_FOUND,
			},
			Success: false,
		}, err
	}
	return &extint.PhotoResponse{
		Status: &extint.ResponseStatus{
			Code: extint.ResponseStatus_RESPONSE_RECEIVED,
		},
		Success: true,
		Image:   imageData,
	}, err
}

func (service *rpcService) Thumbnail(ctx context.Context, in *extint.ThumbnailRequest) (*extint.ThumbnailResponse, error) {
	f, responseChan := engineProtoManager.CreateChannel(&extint.GatewayWrapper_ThumbnailPathMessage{}, 1)
	defer f()

	_, _, err := engineProtoManager.Write(&extint.GatewayWrapper{
		OneofMessageType: &extint.GatewayWrapper_ThumbnailRequest{
			ThumbnailRequest: in,
		},
	})
	if err != nil {
		return nil, err
	}
	payload, ok := <-responseChan
	if !ok {
		return nil, grpc.Errorf(codes.Internal, "Failed to retrieve message")
	}
	if !payload.GetThumbnailPathMessage().GetSuccess() {
		return &extint.ThumbnailResponse{
			Status: &extint.ResponseStatus{
				Code: extint.ResponseStatus_NOT_FOUND,
			},
			Success: false,
		}, err
	}
	imageData, err := SendImageHelper(payload.GetThumbnailPathMessage().GetFullPath())
	if err != nil {
		return &extint.ThumbnailResponse{
			Status: &extint.ResponseStatus{
				Code: extint.ResponseStatus_NOT_FOUND,
			},
			Success: false,
		}, err
	}
	return &extint.ThumbnailResponse{
		Status: &extint.ResponseStatus{
			Code: extint.ResponseStatus_RESPONSE_RECEIVED,
		},
		Success: true,
		Image:   imageData,
	}, err
}

func (service *rpcService) DeletePhoto(ctx context.Context, in *extint.DeletePhotoRequest) (*extint.DeletePhotoResponse, error) {
	f, responseChan := engineProtoManager.CreateChannel(&extint.GatewayWrapper_DeletePhotoResponse{}, 1)
	defer f()

	_, _, err := engineProtoManager.Write(&extint.GatewayWrapper{
		OneofMessageType: &extint.GatewayWrapper_DeletePhotoRequest{
			DeletePhotoRequest: in,
		},
	})
	if err != nil {
		return nil, err
	}
	payload, ok := <-responseChan
	if !ok {
		return nil, grpc.Errorf(codes.Internal, "Failed to retrieve message")
	}
	photoResponse := payload.GetDeletePhotoResponse()
	photoResponse.Status = &extint.ResponseStatus{
		Code: extint.ResponseStatus_RESPONSE_RECEIVED,
	}
	return photoResponse, nil
}

func (service *rpcService) GetLatestAttentionTransfer(ctx context.Context, in *extint.LatestAttentionTransferRequest) (*extint.LatestAttentionTransferResponse, error) {
	f, responseChan := engineProtoManager.CreateChannel(&extint.GatewayWrapper_LatestAttentionTransfer{}, 1)
	defer f()
	_, _, err := engineProtoManager.Write(&extint.GatewayWrapper{
		OneofMessageType: &extint.GatewayWrapper_LatestAttentionTransferRequest{
			LatestAttentionTransferRequest: in,
		},
	})
	if err != nil {
		return nil, err
	}
	attentionTransfer, ok := <-responseChan
	if !ok {
		return nil, grpc.Errorf(codes.Internal, "Failed to retrieve message")
	}
	return &extint.LatestAttentionTransferResponse{
		Status: &extint.ResponseStatus{
			Code: extint.ResponseStatus_RESPONSE_RECEIVED,
		},
		LatestAttentionTransfer: attentionTransfer.GetLatestAttentionTransfer(),
	}, nil
}

func (service *rpcService) ConnectCube(ctx context.Context, in *extint.ConnectCubeRequest) (*extint.ConnectCubeResponse, error) {
	f, responseChan := engineProtoManager.CreateChannel(&extint.GatewayWrapper_ConnectCubeResponse{}, 1)
	defer f()

	_, _, err := engineProtoManager.Write(&extint.GatewayWrapper{
		OneofMessageType: &extint.GatewayWrapper_ConnectCubeRequest{
			ConnectCubeRequest: in,
		},
	})
	if err != nil {
		return nil, err
	}
	gatewayWrapper, ok := <-responseChan
	if !ok {
		return nil, grpc.Errorf(codes.Internal, "Failed to retrieve message")
	}
	response := gatewayWrapper.GetConnectCubeResponse()
	response.Status = &extint.ResponseStatus{
		Code: extint.ResponseStatus_RESPONSE_RECEIVED,
	}
	return response, nil
}

func (service *rpcService) DisconnectCube(ctx context.Context, in *extint.DisconnectCubeRequest) (*extint.DisconnectCubeResponse, error) {
	_, _, err := engineProtoManager.Write(&extint.GatewayWrapper{
		OneofMessageType: &extint.GatewayWrapper_DisconnectCubeRequest{
			DisconnectCubeRequest: in,
		},
	})
	if err != nil {
		return nil, err
	}
	return &extint.DisconnectCubeResponse{
		Status: &extint.ResponseStatus{
			Code: extint.ResponseStatus_REQUEST_PROCESSING,
		},
	}, nil
}

func (service *rpcService) CubesAvailable(ctx context.Context, in *extint.CubesAvailableRequest) (*extint.CubesAvailableResponse, error) {
	f, responseChan := engineProtoManager.CreateChannel(&extint.GatewayWrapper_CubesAvailableResponse{}, 1)
	defer f()
	_, _, err := engineProtoManager.Write(&extint.GatewayWrapper{
		OneofMessageType: &extint.GatewayWrapper_CubesAvailableRequest{
			CubesAvailableRequest: in,
		},
	})
	if err != nil {
		return nil, err
	}
	cubesAvailable, ok := <-responseChan
	if !ok {
		return nil, grpc.Errorf(codes.Internal, "Failed to retrieve message")
	}
	response := cubesAvailable.GetCubesAvailableResponse()
	response.Status = &extint.ResponseStatus{
		Code: extint.ResponseStatus_RESPONSE_RECEIVED,
	}
	return response, nil
}

func (service *rpcService) FlashCubeLights(ctx context.Context, in *extint.FlashCubeLightsRequest) (*extint.FlashCubeLightsResponse, error) {
	_, _, err := engineProtoManager.Write(&extint.GatewayWrapper{
		OneofMessageType: &extint.GatewayWrapper_FlashCubeLightsRequest{
			FlashCubeLightsRequest: in,
		},
	})
	if err != nil {
		return nil, err
	}
	return &extint.FlashCubeLightsResponse{
		Status: &extint.ResponseStatus{
			Code: extint.ResponseStatus_REQUEST_PROCESSING,
		},
	}, nil
}

func (service *rpcService) ForgetPreferredCube(ctx context.Context, in *extint.ForgetPreferredCubeRequest) (*extint.ForgetPreferredCubeResponse, error) {
	_, _, err := engineProtoManager.Write(&extint.GatewayWrapper{
		OneofMessageType: &extint.GatewayWrapper_ForgetPreferredCubeRequest{
			ForgetPreferredCubeRequest: in,
		},
	})
	if err != nil {
		return nil, err
	}
	return &extint.ForgetPreferredCubeResponse{
		Status: &extint.ResponseStatus{
			Code: extint.ResponseStatus_REQUEST_PROCESSING,
		},
	}, nil
}

func (service *rpcService) SetPreferredCube(ctx context.Context, in *extint.SetPreferredCubeRequest) (*extint.SetPreferredCubeResponse, error) {
	_, _, err := engineProtoManager.Write(&extint.GatewayWrapper{
		OneofMessageType: &extint.GatewayWrapper_SetPreferredCubeRequest{
			SetPreferredCubeRequest: in,
		},
	})
	if err != nil {
		return nil, err
	}
	return &extint.SetPreferredCubeResponse{
		Status: &extint.ResponseStatus{
			Code: extint.ResponseStatus_REQUEST_PROCESSING,
		},
	}, nil
}

func (service *rpcService) SetCubeLights(ctx context.Context, in *extint.SetCubeLightsRequest) (*extint.SetCubeLightsResponse, error) {
	_, _, err := engineProtoManager.Write(&extint.GatewayWrapper{
		OneofMessageType: &extint.GatewayWrapper_SetCubeLightsRequest{
			SetCubeLightsRequest: in,
		},
	})
	if err != nil {
		return nil, err
	}
	return &extint.SetCubeLightsResponse{
		Status: &extint.ResponseStatus{
			Code: extint.ResponseStatus_REQUEST_PROCESSING,
		},
	}, nil
}

// NOTE: this is removed from external_interface because the result from the robot (size 100) is too large for the domain socket between gateway and engine
func (service *rpcService) RobotStatusHistory(ctx context.Context, in *extint.RobotHistoryRequest) (*extint.RobotHistoryResponse, error) {
	f, responseChan := engineProtoManager.CreateChannel(&extint.GatewayWrapper_RobotHistoryResponse{}, 1)
	defer f()

	_, _, err := engineProtoManager.Write(&extint.GatewayWrapper{
		OneofMessageType: &extint.GatewayWrapper_RobotHistoryRequest{
			RobotHistoryRequest: in,
		},
	})
	if err != nil {
		return nil, err
	}
	response, ok := <-responseChan
	if !ok {
		return nil, grpc.Errorf(codes.Internal, "Failed to retrieve message")
	}
	return response.GetRobotHistoryResponse(), nil
}

func (service *rpcService) PullJdocs(ctx context.Context, in *extint.PullJdocsRequest) (*extint.PullJdocsResponse, error) {
	f, responseChan := engineProtoManager.CreateChannel(&extint.GatewayWrapper_PullJdocsResponse{}, 1)
	defer f()

	_, _, err := engineProtoManager.Write(&extint.GatewayWrapper{
		OneofMessageType: &extint.GatewayWrapper_PullJdocsRequest{
			PullJdocsRequest: in,
		},
	})
	if err != nil {
		return nil, err
	}
	response, ok := <-responseChan
	if !ok {
		return nil, grpc.Errorf(codes.Internal, "Failed to retrieve message")
	}
	return response.GetPullJdocsResponse(), nil
}

func (service *rpcService) UpdateSettings(ctx context.Context, in *extint.UpdateSettingsRequest) (*extint.UpdateSettingsResponse, error) {
	f, responseChan, ok := engineProtoManager.CreateUniqueChannel(&extint.GatewayWrapper_UpdateSettingsResponse{}, 1)
	if !ok {
		return &extint.UpdateSettingsResponse{
			Status: &extint.ResponseStatus{
				Code: extint.ResponseStatus_ERROR_UPDATE_IN_PROGRESS,
			},
		}, nil
	}
	defer f()

	_, _, err := engineProtoManager.Write(&extint.GatewayWrapper{
		OneofMessageType: &extint.GatewayWrapper_UpdateSettingsRequest{
			UpdateSettingsRequest: in,
		},
	})
	if err != nil {
		return nil, err
	}
	response, ok := <-responseChan
	if !ok {
		return nil, grpc.Errorf(codes.Internal, "Failed to retrieve message")
	}
	return response.GetUpdateSettingsResponse(), nil
}

func (service *rpcService) UpdateAccountSettings(ctx context.Context, in *extint.UpdateAccountSettingsRequest) (*extint.UpdateAccountSettingsResponse, error) {
	f, responseChan, ok := engineProtoManager.CreateUniqueChannel(&extint.GatewayWrapper_UpdateAccountSettingsResponse{}, 1)
	if !ok {
		return &extint.UpdateAccountSettingsResponse{
			Status: &extint.ResponseStatus{
				Code: extint.ResponseStatus_ERROR_UPDATE_IN_PROGRESS,
			},
		}, nil
	}
	defer f()

	_, _, err := engineProtoManager.Write(&extint.GatewayWrapper{
		OneofMessageType: &extint.GatewayWrapper_UpdateAccountSettingsRequest{
			UpdateAccountSettingsRequest: in,
		},
	})
	if err != nil {
		return nil, err
	}
	response, ok := <-responseChan
	if !ok {
		return nil, grpc.Errorf(codes.Internal, "Failed to retrieve message")
	}
	return response.GetUpdateAccountSettingsResponse(), nil
}

func (service *rpcService) UpdateUserEntitlements(ctx context.Context, in *extint.UpdateUserEntitlementsRequest) (*extint.UpdateUserEntitlementsResponse, error) {
	f, responseChan, ok := engineProtoManager.CreateUniqueChannel(&extint.GatewayWrapper_UpdateUserEntitlementsResponse{}, 1)
	if !ok {
		return &extint.UpdateUserEntitlementsResponse{
			Status: &extint.ResponseStatus{
				Code: extint.ResponseStatus_ERROR_UPDATE_IN_PROGRESS,
			},
		}, nil
	}
	defer f()

	_, _, err := engineProtoManager.Write(&extint.GatewayWrapper{
		OneofMessageType: &extint.GatewayWrapper_UpdateUserEntitlementsRequest{
			UpdateUserEntitlementsRequest: in,
		},
	})
	if err != nil {
		return nil, err
	}
	response, ok := <-responseChan
	if !ok {
		return nil, grpc.Errorf(codes.Internal, "Failed to retrieve message")
	}
	return response.GetUpdateUserEntitlementsResponse(), nil
}

// NOTE: this is the only function that won't need to check the client_token_guid header
func (service *rpcService) UserAuthentication(ctx context.Context, in *extint.UserAuthenticationRequest) (*extint.UserAuthenticationResponse, error) {
	if !IsOnRobot {
		return nil, grpc.Errorf(codes.Internal, "User authentication is only available on the robot")
	}

	if !userAuthLimiter.Allow() {
		return nil, grpc.Errorf(codes.ResourceExhausted, "Maximum auth rate exceeded. Please wait and try again later.")
	}

	f, authChan := switchboardManager.CreateChannel(gw_clad.SwitchboardResponseTag_AuthResponse, 1)
	defer f()

	// cap ClientName to 64-characters
	clientName := string(in.ClientName)
	if len(clientName) > 64 {
		clientName = clientName[:64]
	}

	switchboardManager.Write(gw_clad.NewSwitchboardRequestWithAuthRequest(&cloud_clad.AuthRequest{
		SessionToken: string(in.UserSessionId),
		ClientName:   clientName,
		AppId:        "SDK",
	}))
	response, ok := <-authChan
	if !ok {
		return nil, grpc.Errorf(codes.Internal, "Failed to retrieve message")
	}
	auth := response.GetAuthResponse()
	code := extint.UserAuthenticationResponse_UNAUTHORIZED
	token := auth.AppToken
	if auth.Error == cloud_clad.TokenError_NoError {
		code = extint.UserAuthenticationResponse_AUTHORIZED

		// Force an update of the tokens
		response := make(chan struct{})
		tokenManager.ForceUpdate(response)
		<-response
		log.Das("sdk.activate", &log.DasFields{})
	} else {
		token = ""
	}
	return &extint.UserAuthenticationResponse{
		Status: &extint.ResponseStatus{
			Code: extint.ResponseStatus_RESPONSE_RECEIVED,
		},
		Code:            code,
		ClientTokenGuid: []byte(token),
	}, nil
}

func ValidateActionTag(idTag int32) error {
	firstTag := int32(extint.ActionTagConstants_FIRST_SDK_TAG)
	lastTag := int32(extint.ActionTagConstants_LAST_SDK_TAG)
	if idTag < firstTag || idTag > lastTag {
		return grpc.Errorf(codes.InvalidArgument, "Invalid Action tag_id")
	}

	return nil
}

func (service *rpcService) GoToPose(ctx context.Context, in *extint.GoToPoseRequest) (*extint.GoToPoseResponse, error) {

	if err := ValidateActionTag(in.IdTag); err != nil {
		return nil, err
	}

	f, responseChan := engineProtoManager.CreateChannel(&extint.GatewayWrapper_GoToPoseResponse{}, 1)
	defer f()

	_, _, err := engineProtoManager.Write(&extint.GatewayWrapper{
		OneofMessageType: &extint.GatewayWrapper_GoToPoseRequest{
			GoToPoseRequest: in,
		},
	})
	if err != nil {
		return nil, err
	}

	goToPoseResponse, ok := <-responseChan
	if !ok {
		return nil, grpc.Errorf(codes.Internal, "Failed to retrieve message")
	}
	response := goToPoseResponse.GetGoToPoseResponse()
	response.Status = &extint.ResponseStatus{
		Code: extint.ResponseStatus_RESPONSE_RECEIVED,
	}
	log.Printf("Received rpc response GoToPose(%#v)\n", in)
	return response, nil
}

func (service *rpcService) DockWithCube(ctx context.Context, in *extint.DockWithCubeRequest) (*extint.DockWithCubeResponse, error) {

	if err := ValidateActionTag(in.IdTag); err != nil {
		return nil, err
	}

	f, responseChan := engineProtoManager.CreateChannel(&extint.GatewayWrapper_DockWithCubeResponse{}, 1)
	defer f()

	_, _, err := engineProtoManager.Write(&extint.GatewayWrapper{
		OneofMessageType: &extint.GatewayWrapper_DockWithCubeRequest{
			DockWithCubeRequest: in,
		},
	})
	if err != nil {
		return nil, err
	}

	dockWithCubeResponse, ok := <-responseChan
	if !ok {
		return nil, grpc.Errorf(codes.Internal, "Failed to retrieve message")
	}
	response := dockWithCubeResponse.GetDockWithCubeResponse()
	response.Status = &extint.ResponseStatus{
		Code: extint.ResponseStatus_RESPONSE_RECEIVED,
	}
	return response, nil
}

func (service *rpcService) DriveStraight(ctx context.Context, in *extint.DriveStraightRequest) (*extint.DriveStraightResponse, error) {

	if err := ValidateActionTag(in.IdTag); err != nil {
		return nil, err
	}

	f, responseChan := engineProtoManager.CreateChannel(&extint.GatewayWrapper_DriveStraightResponse{}, 1)
	defer f()

	_, _, err := engineProtoManager.Write(&extint.GatewayWrapper{
		OneofMessageType: &extint.GatewayWrapper_DriveStraightRequest{
			DriveStraightRequest: in,
		},
	})
	if err != nil {
		return nil, err
	}

	driveStraightResponse, ok := <-responseChan
	if !ok {
		return nil, grpc.Errorf(codes.Internal, "Failed to retrieve message")
	}
	response := driveStraightResponse.GetDriveStraightResponse()
	response.Status = &extint.ResponseStatus{
		Code: extint.ResponseStatus_RESPONSE_RECEIVED,
	}
	return response, nil
}

func (service *rpcService) TurnInPlace(ctx context.Context, in *extint.TurnInPlaceRequest) (*extint.TurnInPlaceResponse, error) {

	if err := ValidateActionTag(in.IdTag); err != nil {
		return nil, err
	}

	f, responseChan := engineProtoManager.CreateChannel(&extint.GatewayWrapper_TurnInPlaceResponse{}, 1)
	defer f()

	_, _, err := engineProtoManager.Write(&extint.GatewayWrapper{
		OneofMessageType: &extint.GatewayWrapper_TurnInPlaceRequest{
			TurnInPlaceRequest: in,
		},
	})
	if err != nil {
		return nil, err
	}

	turnInPlaceResponse, ok := <-responseChan
	if !ok {
		return nil, grpc.Errorf(codes.Internal, "Failed to retrieve message")
	}
	response := turnInPlaceResponse.GetTurnInPlaceResponse()
	response.Status = &extint.ResponseStatus{
		Code: extint.ResponseStatus_RESPONSE_RECEIVED,
	}
	return response, nil
}

func (service *rpcService) SetHeadAngle(ctx context.Context, in *extint.SetHeadAngleRequest) (*extint.SetHeadAngleResponse, error) {

	if err := ValidateActionTag(in.IdTag); err != nil {
		return nil, err
	}

	f, responseChan := engineProtoManager.CreateChannel(&extint.GatewayWrapper_SetHeadAngleResponse{}, 1)
	defer f()

	_, _, err := engineProtoManager.Write(&extint.GatewayWrapper{
		OneofMessageType: &extint.GatewayWrapper_SetHeadAngleRequest{
			SetHeadAngleRequest: in,
		},
	})
	if err != nil {
		return nil, err
	}

	setHeadAngleResponse, ok := <-responseChan
	if !ok {
		return nil, grpc.Errorf(codes.Internal, "Failed to retrieve message")
	}
	response := setHeadAngleResponse.GetSetHeadAngleResponse()
	response.Status = &extint.ResponseStatus{
		Code: extint.ResponseStatus_RESPONSE_RECEIVED,
	}
	return response, nil
}

func (service *rpcService) SetLiftHeight(ctx context.Context, in *extint.SetLiftHeightRequest) (*extint.SetLiftHeightResponse, error) {

	if err := ValidateActionTag(in.IdTag); err != nil {
		return nil, err
	}

	f, responseChan := engineProtoManager.CreateChannel(&extint.GatewayWrapper_SetLiftHeightResponse{}, 1)
	defer f()

	_, _, err := engineProtoManager.Write(&extint.GatewayWrapper{
		OneofMessageType: &extint.GatewayWrapper_SetLiftHeightRequest{
			SetLiftHeightRequest: in,
		},
	})
	if err != nil {
		return nil, err
	}

	setLiftHeightResponse, ok := <-responseChan
	if !ok {
		return nil, grpc.Errorf(codes.Internal, "Failed to retrieve message")
	}
	response := setLiftHeightResponse.GetSetLiftHeightResponse()
	response.Status = &extint.ResponseStatus{
		Code: extint.ResponseStatus_RESPONSE_RECEIVED,
	}
	return response, nil
}

func (service *rpcService) TurnTowardsFace(ctx context.Context, in *extint.TurnTowardsFaceRequest) (*extint.TurnTowardsFaceResponse, error) {

	if err := ValidateActionTag(in.IdTag); err != nil {
		return nil, err
	}

	f, responseChan := engineProtoManager.CreateChannel(&extint.GatewayWrapper_TurnTowardsFaceResponse{}, 1)
	defer f()

	_, _, err := engineProtoManager.Write(&extint.GatewayWrapper{
		OneofMessageType: &extint.GatewayWrapper_TurnTowardsFaceRequest{
			TurnTowardsFaceRequest: in,
		},
	})
	if err != nil {
		return nil, err
	}

	turnTowardsFaceResponse, ok := <-responseChan
	if !ok {
		return nil, grpc.Errorf(codes.Internal, "Failed to retrieve message")
	}
	response := turnTowardsFaceResponse.GetTurnTowardsFaceResponse()
	response.Status = &extint.ResponseStatus{
		Code: extint.ResponseStatus_RESPONSE_RECEIVED,
	}
	return response, nil
}

func (service *rpcService) GoToObject(ctx context.Context, in *extint.GoToObjectRequest) (*extint.GoToObjectResponse, error) {

	if err := ValidateActionTag(in.IdTag); err != nil {
		return nil, err
	}

	f, responseChan := engineProtoManager.CreateChannel(&extint.GatewayWrapper_GoToObjectResponse{}, 1)
	defer f()

	_, _, err := engineProtoManager.Write(&extint.GatewayWrapper{
		OneofMessageType: &extint.GatewayWrapper_GoToObjectRequest{
			GoToObjectRequest: in,
		},
	})
	if err != nil {
		return nil, err
	}

	goToObjectResponse, ok := <-responseChan
	if !ok {
		return nil, grpc.Errorf(codes.Internal, "Failed to retrieve message")
	}
	response := goToObjectResponse.GetGoToObjectResponse()
	response.Status = &extint.ResponseStatus{
		Code: extint.ResponseStatus_RESPONSE_RECEIVED,
	}
	return response, nil
}

func (service *rpcService) RollObject(ctx context.Context, in *extint.RollObjectRequest) (*extint.RollObjectResponse, error) {

	if err := ValidateActionTag(in.IdTag); err != nil {
		return nil, err
	}

	f, responseChan := engineProtoManager.CreateChannel(&extint.GatewayWrapper_RollObjectResponse{}, 1)
	defer f()

	_, _, err := engineProtoManager.Write(&extint.GatewayWrapper{
		OneofMessageType: &extint.GatewayWrapper_RollObjectRequest{
			RollObjectRequest: in,
		},
	})
	if err != nil {
		return nil, err
	}

	rollObjectResponse, ok := <-responseChan
	if !ok {
		return nil, grpc.Errorf(codes.Internal, "Failed to retrieve message")
	}
	response := rollObjectResponse.GetRollObjectResponse()
	response.Status = &extint.ResponseStatus{
		Code: extint.ResponseStatus_RESPONSE_RECEIVED,
	}
	return response, nil
}

func (service *rpcService) PopAWheelie(ctx context.Context, in *extint.PopAWheelieRequest) (*extint.PopAWheelieResponse, error) {

	if err := ValidateActionTag(in.IdTag); err != nil {
		return nil, err
	}

	f, responseChan := engineProtoManager.CreateChannel(&extint.GatewayWrapper_PopAWheelieResponse{}, 1)
	defer f()

	_, _, err := engineProtoManager.Write(&extint.GatewayWrapper{
		OneofMessageType: &extint.GatewayWrapper_PopAWheelieRequest{
			PopAWheelieRequest: in,
		},
	})
	if err != nil {
		return nil, err
	}

	popAWheelieResponse, ok := <-responseChan
	if !ok {
		return nil, grpc.Errorf(codes.Internal, "Failed to retrieve message")
	}
	response := popAWheelieResponse.GetPopAWheelieResponse()
	response.Status = &extint.ResponseStatus{
		Code: extint.ResponseStatus_RESPONSE_RECEIVED,
	}
	return response, nil
}

func (service *rpcService) PickupObject(ctx context.Context, in *extint.PickupObjectRequest) (*extint.PickupObjectResponse, error) {

	if err := ValidateActionTag(in.IdTag); err != nil {
		return nil, err
	}

	f, responseChan := engineProtoManager.CreateChannel(&extint.GatewayWrapper_PickupObjectResponse{}, 1)
	defer f()

	_, _, err := engineProtoManager.Write(&extint.GatewayWrapper{
		OneofMessageType: &extint.GatewayWrapper_PickupObjectRequest{
			PickupObjectRequest: in,
		},
	})
	if err != nil {
		return nil, err
	}

	pickupObjectResponse, ok := <-responseChan
	if !ok {
		return nil, grpc.Errorf(codes.Internal, "Failed to retrieve message")
	}
	response := pickupObjectResponse.GetPickupObjectResponse()
	response.Status = &extint.ResponseStatus{
		Code: extint.ResponseStatus_RESPONSE_RECEIVED,
	}
	return response, nil
}

func (service *rpcService) PlaceObjectOnGroundHere(ctx context.Context, in *extint.PlaceObjectOnGroundHereRequest) (*extint.PlaceObjectOnGroundHereResponse, error) {

	if err := ValidateActionTag(in.IdTag); err != nil {
		return nil, err
	}

	f, responseChan := engineProtoManager.CreateChannel(&extint.GatewayWrapper_PlaceObjectOnGroundHereResponse{}, 1)
	defer f()

	_, _, err := engineProtoManager.Write(&extint.GatewayWrapper{
		OneofMessageType: &extint.GatewayWrapper_PlaceObjectOnGroundHereRequest{
			PlaceObjectOnGroundHereRequest: in,
		},
	})
	if err != nil {
		return nil, err
	}

	placeObjectOnGroundResponse, ok := <-responseChan
	if !ok {
		return nil, grpc.Errorf(codes.Internal, "Failed to retrieve message")
	}
	response := placeObjectOnGroundResponse.GetPlaceObjectOnGroundHereResponse()
	response.Status = &extint.ResponseStatus{
		Code: extint.ResponseStatus_RESPONSE_RECEIVED,
	}
	return response, nil
}

func (service *rpcService) SetMasterVolume(ctx context.Context, in *extint.MasterVolumeRequest) (*extint.MasterVolumeResponse, error) {
	f, responseChan := engineProtoManager.CreateChannel(&extint.GatewayWrapper_MasterVolumeResponse{}, 1)
	defer f()

	_, _, err := engineProtoManager.Write(&extint.GatewayWrapper{
		OneofMessageType: &extint.GatewayWrapper_MasterVolumeRequest{
			MasterVolumeRequest: in,
		},
	})
	if err != nil {
		return nil, err
	}

	masterVolumeResponse, ok := <-responseChan
	if !ok {
		return nil, grpc.Errorf(codes.Internal, "Failed to retrieve message")
	}
	response := masterVolumeResponse.GetMasterVolumeResponse()
	response.Status = &extint.ResponseStatus{
		Code: extint.ResponseStatus_RESPONSE_RECEIVED,
	}
	return response, nil
}

func (service *rpcService) BatteryState(ctx context.Context, in *extint.BatteryStateRequest) (*extint.BatteryStateResponse, error) {
	f, responseChan := engineProtoManager.CreateChannel(&extint.GatewayWrapper_BatteryStateResponse{}, 1)
	defer f()

	_, _, err := engineProtoManager.Write(&extint.GatewayWrapper{
		OneofMessageType: &extint.GatewayWrapper_BatteryStateRequest{
			BatteryStateRequest: in,
		},
	})
	if err != nil {
		return nil, err
	}
	payload, ok := <-responseChan
	if !ok {
		return nil, grpc.Errorf(codes.Internal, "Failed to retrieve message")
	}
	payload.GetBatteryStateResponse().Status = &extint.ResponseStatus{
		Code: extint.ResponseStatus_RESPONSE_RECEIVED,
	}
	return payload.GetBatteryStateResponse(), nil
}

func (service *rpcService) VersionState(ctx context.Context, in *extint.VersionStateRequest) (*extint.VersionStateResponse, error) {
	f, responseChan := engineProtoManager.CreateChannel(&extint.GatewayWrapper_VersionStateResponse{}, 1)
	defer f()

	_, _, err := engineProtoManager.Write(&extint.GatewayWrapper{
		OneofMessageType: &extint.GatewayWrapper_VersionStateRequest{
			VersionStateRequest: in,
		},
	})
	if err != nil {
		return nil, err
	}
	payload, ok := <-responseChan
	if !ok {
		return nil, grpc.Errorf(codes.Internal, "Failed to retrieve message")
	}
	payload.GetVersionStateResponse().Status = &extint.ResponseStatus{
		Code: extint.ResponseStatus_RESPONSE_RECEIVED,
	}
	return payload.GetVersionStateResponse(), nil
}

func (service *rpcService) SayText(ctx context.Context, in *extint.SayTextRequest) (*extint.SayTextResponse, error) {
	f, responseChan := engineProtoManager.CreateChannel(&extint.GatewayWrapper_SayTextResponse{}, 1)
	defer f()

	_, _, err := engineProtoManager.Write(&extint.GatewayWrapper{
		OneofMessageType: &extint.GatewayWrapper_SayTextRequest{
			SayTextRequest: in,
		},
	})
	if err != nil {
		return nil, err
	}
	done := false
	var sayTextResponse *extint.SayTextResponse
	for !done {
		payload, ok := <-responseChan
		if !ok {
			return nil, grpc.Errorf(codes.Internal, "Failed to retrieve message")
		}
		sayTextResponse = payload.GetSayTextResponse()
		state := sayTextResponse.GetState()
		if state == extint.SayTextResponse_FINISHED {
			done = true
		} else if state == extint.SayTextResponse_INVALID {
			return nil, grpc.Errorf(codes.Internal, "Failed to say text")
		}
	}
	sayTextResponse.Status = &extint.ResponseStatus{
		Code: extint.ResponseStatus_RESPONSE_RECEIVED,
	}
	return sayTextResponse, nil
}

func AudioSendModeRequest(mode extint.AudioProcessingMode) error {
	log.Println("SDK Requesting Audio with mode(", mode, ")")

	_, _, err := engineProtoManager.Write(&extint.GatewayWrapper{
		OneofMessageType: &extint.GatewayWrapper_AudioSendModeRequest{
			AudioSendModeRequest: &extint.AudioSendModeRequest{
				Mode: mode,
			},
		},
	})

	return err
}

type AudioFeedCache struct {
	Data        []byte
	GroupId     int32
	Invalid     bool
	Size        int32
	LastChunkId int32
}

func ResetAudioCache(cache *AudioFeedCache) error {
	cache.Data = nil
	cache.GroupId = -1
	cache.Invalid = false
	cache.Size = 0
	cache.LastChunkId = -1
	return nil
}

func UnpackAudioChunk(audioChunk *extint.AudioChunk, cache *AudioFeedCache) bool {
	groupId := int32(audioChunk.GetGroupId())
	chunkId := int32(audioChunk.GetChunkId())

	if cache.GroupId != -1 && chunkId == 0 {
		if !cache.Invalid {
			log.Errorln("Lost final chunk of audio group; discarding")
		}
		cache.GroupId = -1
	}

	if cache.GroupId == -1 {
		if chunkId != 0 {
			if !cache.Invalid {
				log.Errorln("Received chunk of broken audio stream")
			}
			cache.Invalid = true
			return false
		}
		// discard any previous in-progress image
		ResetAudioCache(cache)

		cache.Data = make([]byte, extint.AudioConstants_SAMPLE_COUNTS_PER_SDK_MESSAGE*2)
		cache.GroupId = int32(groupId)
	}

	if chunkId != cache.LastChunkId+1 || groupId != cache.GroupId {
		log.Errorf("Audio missing chunks; discarding (last_chunk_id=%i partial_audio_group_id=%i)\n", cache.LastChunkId, cache.GroupId)
		ResetAudioCache(cache)
		cache.Invalid = true
		return false
	}

	dataSize := int32(len(audioChunk.GetSignalPower()))
	copy(cache.Data[cache.Size:cache.Size+dataSize], audioChunk.GetSignalPower()[:])
	cache.Size += dataSize
	cache.LastChunkId = chunkId

	return chunkId == int32(audioChunk.GetAudioChunkCount()-1)
}

// Long running message for sending audio feed to listening sdk users
func (service *rpcService) AudioFeed(in *extint.AudioFeedRequest, stream extint.ExternalInterface_AudioFeedServer) error {
	// @TODO: Expose other audio processing modes
	//
	// The composite multi-microphone non-beamforming (AUDIO_VOICE_DETECT_MODE) mode has been identified as the best for voice detection,
	// as well as incidentally calculating directional and noise_floor data.  As such it's most reasonable as the SDK's default mode.
	//
	// While this mode will send directional source data, it is different from DIRECTIONAL_MODE in that it does not isolate and clean
	// up the sound stream with respect to the loudest direction (which makes the result more human-ear pleasing but more ml difficult).
	//
	// It should however be noted that the robot will automatically shift into FAST_MODE (cleaned up single microphone) when
	// entering low power mode, so its important to leave this exposed.
	//

	// Enable audio stream
	err := AudioSendModeRequest(extint.AudioProcessingMode_AUDIO_VOICE_DETECT_MODE)
	if err != nil {
		return err
	}

	// Disable audio stream
	defer AudioSendModeRequest(extint.AudioProcessingMode_AUDIO_OFF)

	// Forward audio data from engine
	f, audioFeedChannel := engineProtoManager.CreateChannel(&extint.GatewayWrapper_AudioChunk{}, 1024)
	defer f()

	cache := AudioFeedCache{
		Data:    nil,
		GroupId: -1,
		Invalid: false,
		Size:    0,
	}

	for result := range audioFeedChannel {

		audioChunk := result.GetAudioChunk()

		readyToSend := UnpackAudioChunk(audioChunk, &cache)
		if readyToSend {
			audioFeedResponse := &extint.AudioFeedResponse{
				RobotTimeStamp:     audioChunk.GetRobotTimeStamp(),
				GroupId:            uint32(audioChunk.GetGroupId()),
				SignalPower:        cache.Data[0:cache.Size],
				DirectionStrengths: audioChunk.GetDirectionStrengths(),
				SourceDirection:    audioChunk.GetSourceDirection(),
				SourceConfidence:   audioChunk.GetSourceConfidence(),
				NoiseFloorPower:    audioChunk.GetNoiseFloorPower(),
			}
			ResetAudioCache(&cache)

			if err := stream.Send(audioFeedResponse); err != nil {
				return err
			} else if err = stream.Context().Err(); err != nil {
				// This is the case where the user disconnects the stream
				// We should still return the err in case the user doesn't think they disconnected
				return err
			}
		}
	}

	errMsg := "AudioChunk engine stream died unexpectedly"
	log.Errorln(errMsg)
	return grpc.Errorf(codes.Internal, errMsg)
}

func (service *rpcService) EnableMarkerDetection(ctx context.Context, request *extint.EnableMarkerDetectionRequest) (*extint.EnableMarkerDetectionResponse, error) {
	f, responseChan := engineProtoManager.CreateChannel(&extint.GatewayWrapper_EnableMarkerDetectionResponse{}, 1)
	defer f()

	_, _, err := engineProtoManager.Write(&extint.GatewayWrapper{
		OneofMessageType: &extint.GatewayWrapper_EnableMarkerDetectionRequest{
			EnableMarkerDetectionRequest: request,
		},
	})
	if err != nil {
		return nil, err
	}

	if request.Enable == false {
		// VIC-12762 EnableMarkerDetectionRequest is requesting that the marker detection vision mode be turned off.
		// There are cases when there is another subscriber on the robot (not affiliated with the SDK)
		// that wants to keep this vision mode on. So this request to turn it off is more of a suggestion.
		// Don't wait for the response to come back from the vision system via sdkComponent to confirm that
		// this request was successful.
		//
		// This fixes the problem where the SDK either hangs or waits ~50 seconds intermittently for confirmation
		// that the vision mode was turned off.
		//
		// TODO Do other vision modes encounter the same problem?
		//
		// Create response, send it and return early.
		return &extint.EnableMarkerDetectionResponse{
			Status: &extint.ResponseStatus{
				Code: extint.ResponseStatus_REQUEST_PROCESSING,
			},
		}, nil
	} else {
		payload, ok := <-responseChan
		if !ok {
			return nil, grpc.Errorf(codes.Internal, "Failed to retrieve message")
		}

		response := payload.GetEnableMarkerDetectionResponse()
		response.Status = &extint.ResponseStatus{
			Code: extint.ResponseStatus_RESPONSE_RECEIVED,
		}

		return response, nil
	}
}

func (service *rpcService) EnableFaceDetection(ctx context.Context, request *extint.EnableFaceDetectionRequest) (*extint.EnableFaceDetectionResponse, error) {
	f, responseChan := engineProtoManager.CreateChannel(&extint.GatewayWrapper_EnableFaceDetectionResponse{}, 1)
	defer f()

	_, _, err := engineProtoManager.Write(&extint.GatewayWrapper{
		OneofMessageType: &extint.GatewayWrapper_EnableFaceDetectionRequest{
			EnableFaceDetectionRequest: request,
		},
	})
	if err != nil {
		return nil, err
	}

	payload, ok := <-responseChan
	if !ok {
		return nil, grpc.Errorf(codes.Internal, "Failed to retrieve message")
	}

	response := payload.GetEnableFaceDetectionResponse()
	response.Status = &extint.ResponseStatus{
		Code: extint.ResponseStatus_RESPONSE_RECEIVED,
	}

	return response, nil
}

func (service *rpcService) EnableMotionDetection(ctx context.Context, request *extint.EnableMotionDetectionRequest) (*extint.EnableMotionDetectionResponse, error) {
	f, responseChan := engineProtoManager.CreateChannel(&extint.GatewayWrapper_EnableMotionDetectionResponse{}, 1)
	defer f()

	_, _, err := engineProtoManager.Write(&extint.GatewayWrapper{
		OneofMessageType: &extint.GatewayWrapper_EnableMotionDetectionRequest{
			EnableMotionDetectionRequest: request,
		},
	})
	if err != nil {
		return nil, err
	}

	payload, ok := <-responseChan
	if !ok {
		return nil, grpc.Errorf(codes.Internal, "Failed to retrieve message")
	}

	response := payload.GetEnableMotionDetectionResponse()
	response.Status = &extint.ResponseStatus{
		Code: extint.ResponseStatus_RESPONSE_RECEIVED,
	}

	return response, nil
}

func (service *rpcService) EnableMirrorMode(ctx context.Context, request *extint.EnableMirrorModeRequest) (*extint.EnableMirrorModeResponse, error) {
	f, responseChan := engineProtoManager.CreateChannel(&extint.GatewayWrapper_EnableMirrorModeResponse{}, 1)
	defer f()

	_, _, err := engineProtoManager.Write(&extint.GatewayWrapper{
		OneofMessageType: &extint.GatewayWrapper_EnableMirrorModeRequest{
			EnableMirrorModeRequest: request,
		},
	})
	if err != nil {
		return nil, err
	}

	payload, ok := <-responseChan
	if !ok {
		return nil, grpc.Errorf(codes.Internal, "Failed to retrieve message")
	}

	response := payload.GetEnableMirrorModeResponse()
	response.Status = &extint.ResponseStatus{
		Code: extint.ResponseStatus_RESPONSE_RECEIVED,
	}

	return response, nil
}

// Capture a single image using the camera
func (service *rpcService) CaptureSingleImage(ctx context.Context, request *extint.CaptureSingleImageRequest) (*extint.CaptureSingleImageResponse, error) {
	// Enable image stream
	_, err := service.EnableImageStreaming(nil, &extint.EnableImageStreamingRequest{
		Enable:               true,
		EnableHighResolution: request.EnableHighResolution,
	})

	if err != nil {
		return nil, err
	}

	// Disable image stream
	defer service.EnableImageStreaming(nil, &extint.EnableImageStreamingRequest{
		Enable:               false,
		EnableHighResolution: false,
	})

	f, cameraFeedChannel := engineProtoManager.CreateChannel(&extint.GatewayWrapper_ImageChunk{}, 1024)
	defer f()

	cache := CameraFeedCache{
		Data:    nil,
		ImageId: -1,
		Invalid: false,
		Size:    0,
	}

	for result := range cameraFeedChannel {
		imageChunk := result.GetImageChunk()
		readyToSend := UnpackCameraImageChunk(imageChunk, &cache)
		if readyToSend {
			capturedSingleImage := &extint.CaptureSingleImageResponse{
				FrameTimeStamp: imageChunk.GetFrameTimeStamp(),
				ImageId:        uint32(cache.ImageId),
				ImageEncoding:  imageChunk.GetImageEncoding(),
				Data:           cache.Data[0:cache.Size],
			}
			capturedSingleImage.Status = &extint.ResponseStatus{Code: extint.ResponseStatus_RESPONSE_RECEIVED}
			return capturedSingleImage, nil
		}
	}

	errMsg := "ImageChunk engine stream died unexpectedly"
	log.Errorln(errMsg)
	return nil, grpc.Errorf(codes.Internal, errMsg)
}

// TODO VIC-11579 Support specifying streaming resolution
func (service *rpcService) EnableImageStreaming(ctx context.Context, request *extint.EnableImageStreamingRequest) (*extint.EnableImageStreamingResponse, error) {
	f, responseChan := engineProtoManager.CreateChannel(&extint.GatewayWrapper_EnableImageStreamingResponse{}, 1)
	defer f()

	_, _, err := engineProtoManager.Write(&extint.GatewayWrapper{
		OneofMessageType: &extint.GatewayWrapper_EnableImageStreamingRequest{
			EnableImageStreamingRequest: request,
		},
	})
	if err != nil {
		return nil, err
	}

	payload, ok := <-responseChan
	if !ok {
		return nil, grpc.Errorf(codes.Internal, "Failed to retrieve message")
	}

	response := payload.GetEnableImageStreamingResponse()
	response.Status = &extint.ResponseStatus{
		Code: extint.ResponseStatus_RESPONSE_RECEIVED,
	}

	return response, nil
}

// indicates if image streaming is enabled or not
func (service *rpcService) IsImageStreamingEnabled(ctx context.Context, request *extint.IsImageStreamingEnabledRequest) (*extint.IsImageStreamingEnabledResponse, error) {
	f, responseChan := engineProtoManager.CreateChannel(&extint.GatewayWrapper_IsImageStreamingEnabledResponse{}, 1)
	defer f()
	_, _, err := engineProtoManager.Write(&extint.GatewayWrapper{
		OneofMessageType: &extint.GatewayWrapper_IsImageStreamingEnabledRequest{
			IsImageStreamingEnabledRequest: request,
		},
	})
	if err != nil {
		return nil, err
	}
	payload, ok := <-responseChan
	if !ok {
		return nil, grpc.Errorf(codes.Internal, "Failed to retrieve message")
	}
	response := payload.GetIsImageStreamingEnabledResponse()
	return &extint.IsImageStreamingEnabledResponse{
		IsImageStreamingEnabled: response.IsImageStreamingEnabled,
	}, nil
}

type CameraFeedCache struct {
	Data        []byte
	ImageId     int32
	Invalid     bool
	Size        int32
	LastChunkId int32
}

func ResetCameraCache(cache *CameraFeedCache) error {
	cache.Data = nil
	cache.ImageId = -1
	cache.Invalid = false
	cache.Size = 0
	cache.LastChunkId = -1
	return nil
}

func UnpackCameraImageChunk(imageChunk *extint.ImageChunk, cache *CameraFeedCache) bool {
	imageId := int32(imageChunk.GetImageId())
	chunkId := int32(imageChunk.GetChunkId())

	if cache.ImageId != -1 && chunkId == 0 {
		if !cache.Invalid {
			log.Errorln("Lost final chunk of image; discarding")
		}
		cache.ImageId = -1
	}

	if cache.ImageId == -1 {
		if chunkId != 0 {
			if !cache.Invalid {
				log.Errorln("Received chunk of broken image")
			}
			cache.Invalid = true
			return false
		}
		// discard any previous in-progress image
		ResetCameraCache(cache)

		cache.Data = make([]byte, imageChunk.GetWidth()*imageChunk.GetHeight()*3)
		cache.ImageId = int32(imageId)
	}

	if chunkId != cache.LastChunkId+1 || imageId != cache.ImageId {
		log.Errorf("Image missing chunks; discarding (last_chunk_id=%i partial_image_id=%i)\n", cache.LastChunkId, cache.ImageId)
		ResetCameraCache(cache)
		cache.Invalid = true
		return false
	}

	dataSize := int32(len(imageChunk.GetData()))
	copy(cache.Data[cache.Size:cache.Size+dataSize], imageChunk.GetData()[:])
	cache.Size += dataSize
	cache.LastChunkId = chunkId

	return chunkId == int32(imageChunk.GetImageChunkCount()-1)
}

// Long running message for sending camera feed to listening sdk users
func (service *rpcService) CameraFeed(in *extint.CameraFeedRequest, stream extint.ExternalInterface_CameraFeedServer) error {
	// Enable video stream. The video stream only uses the default image resolution.
	_, err := service.EnableImageStreaming(nil, &extint.EnableImageStreamingRequest{
		Enable:               true,
		EnableHighResolution: false,
	})

	if err != nil {
		return err
	}

	// Disable video stream
	defer service.EnableImageStreaming(nil, &extint.EnableImageStreamingRequest{
		Enable:               false,
		EnableHighResolution: false,
	})

	f, cameraFeedChannel := engineProtoManager.CreateChannel(&extint.GatewayWrapper_ImageChunk{}, 1024)
	defer f()

	cache := CameraFeedCache{
		Data:    nil,
		ImageId: -1,
		Invalid: false,
		Size:    0,
	}

	for result := range cameraFeedChannel {

		imageChunk := result.GetImageChunk()
		readyToSend := UnpackCameraImageChunk(imageChunk, &cache)
		if readyToSend {
			cameraFeedResponse := &extint.CameraFeedResponse{
				FrameTimeStamp: imageChunk.GetFrameTimeStamp(),
				ImageId:        uint32(cache.ImageId),
				ImageEncoding:  imageChunk.GetImageEncoding(),
				Data:           cache.Data[0:cache.Size],
			}
			ResetCameraCache(&cache)

			if err := stream.Send(cameraFeedResponse); err != nil {
				return err
			} else if err = stream.Context().Err(); err != nil {
				// This is the case where the user disconnects the stream
				// We should still return the err in case the user doesn't think they disconnected
				return err
			}
		}
	}

	errMsg := "ImageChunk engine stream died unexpectedly"
	log.Errorln(errMsg)
	return grpc.Errorf(codes.Internal, errMsg)
}

func ReadStringFromFile(filename string, defaultValue string) string {
	returnValue := defaultValue
	if data, err := ioutil.ReadFile(filename); err == nil {
		returnValue = strings.TrimSpace(string(data))
	}
	return returnValue
}

func ReadInt64FromFile(filename string, defaultValue int64) int64 {
	returnValue := defaultValue
	data := ReadStringFromFile(filename, "")
	if len(data) > 0 {
		if val, err := strconv.ParseInt(data, 0, 64); err == nil {
			returnValue = val
		}
	}
	return returnValue
}

const (
	otaExitCodeNotAvailable  = int64(-1)
	otaExitCodeSuccess       = int64(0)
	otaExitCodeIOError       = int64(208)
	otaExitCodeSocketTimeout = int64(215)
)

// GetUpdateStatus tells if the robot is ready to reboot and update.
func (service *rpcService) GetUpdateStatus() (*extint.CheckUpdateStatusResponse, error) {
	updateStatus := &extint.CheckUpdateStatusResponse{
		Status: &extint.ResponseStatus{
			Code: extint.ResponseStatus_OK,
		},
		UpdateStatus:  extint.CheckUpdateStatusResponse_NO_UPDATE,
		Progress:      ReadInt64FromFile("/run/update-engine/progress", -1),
		Expected:      ReadInt64FromFile("/run/update-engine/expected-size", -1),
		ExitCode:      ReadInt64FromFile("/run/update-engine/exit_code", otaExitCodeNotAvailable),
		Error:         ReadStringFromFile("/run/update-engine/error", ""),
		UpdatePhase:   ReadStringFromFile("/run/update-engine/phase", ""),
		UpdateVersion: "",
	}

	if data, err := ioutil.ReadFile("/run/update-engine/manifest.ini"); err == nil {
		updateVersionExpr := regexp.MustCompile("update_version\\s*=\\s*(\\S*)")
		match := updateVersionExpr.FindStringSubmatch(string(data))
		if len(match) == 2 {
			updateStatus.UpdateVersion = match[1]
		}
	}

	// With one exception, an ExitCode > otaExitCodeSuccess means we encountered an error and need to
	// report it
	if updateStatus.ExitCode > otaExitCodeSuccess {
		updateStatus.UpdateStatus = extint.CheckUpdateStatusResponse_FAILURE_OTHER

		// If we are in the 'download' phase, make our failure status more precise
		// depending on the exit code
		if updateStatus.UpdatePhase == "download" {
			if updateStatus.Error == "Failed to open URL: <urlopen error [Errno -2] Name or service not known>" ||
				updateStatus.ExitCode == otaExitCodeIOError ||
				updateStatus.ExitCode == otaExitCodeSocketTimeout {
				updateStatus.UpdateStatus = extint.CheckUpdateStatusResponse_FAILURE_INTERRUPTED_DOWNLOAD
			}
		}
		// I do this, rather than checking for the 203 because the 203 has other meanings.
		// The below unique error string is what we expect when there is no update available.
		// It is not an indication of failure
		if strings.Contains(updateStatus.Error, "Failed to open URL: HTTP Error 403: Forbidden") {
			updateStatus.UpdateStatus = extint.CheckUpdateStatusResponse_NO_UPDATE
			updateStatus.Error = ""
			updateStatus.ExitCode = otaExitCodeSuccess
		}
		return updateStatus, nil
	}

	// If /run/update-engine/done exists that means that /anki/bin/update-engine
	// successfully downloaded and applied an OS update.  We are now waiting to reboot
	// into it.
	if _, err := os.Stat("/run/update-engine/done"); err == nil {
		updateStatus.UpdateStatus =
			extint.CheckUpdateStatusResponse_READY_TO_REBOOT_INTO_NEW_OS_VERSION
		updateStatus.Error = ""
		updateStatus.ExitCode = otaExitCodeSuccess
		return updateStatus, nil
	}

	// If we don't have an exit code yet, we are in progress
	if updateStatus.ExitCode == otaExitCodeNotAvailable {
		updateStatus.UpdateStatus = extint.CheckUpdateStatusResponse_IN_PROGRESS_STARTING
		// If we don't have an exit code, we should not report an error until we do
		updateStatus.Error = ""
	}

	if updateStatus.Progress > 0 {
		if updateStatus.UpdatePhase == "download" {
			updateStatus.UpdateStatus = extint.CheckUpdateStatusResponse_IN_PROGRESS_DOWNLOAD
		} else {
			updateStatus.UpdateStatus = extint.CheckUpdateStatusResponse_IN_PROGRESS_OTHER
		}
	}

	return updateStatus, nil
}

// UpdateStatusStream tells if the robot is ready to reboot and update.
func (service *rpcService) UpdateStatusStream() {
	updateStarted := false
	// If this co-routine is already running, we don't need another one
	if statusStreamRunning {
		return
	}
	statusStreamRunning = true
	defer func() {
		statusStreamRunning = false
	}()
	iterations := 0
	for len(connectionId) == 0 && iterations < 10 {
		// Wait a bit to be sure that the connectionId is valid before continuing.
		time.Sleep(250 * time.Millisecond)
		iterations++
	}

	for len(connectionId) != 0 {
		status, err := service.GetUpdateStatus()
		// Keep streaming to the requestor until they disconnect. We don't stop
		// streaming just because there's no update pending (a requested update
		// may be pending, but hasn't had a chance to update
		// /run/update-engine/* yet).
		if err != nil {
			break
		}

		// It is possible that we will not have new values for 'progress' and 'expected'
		// or we could be in the middle of a transition where 'expected < progress'. We
		// don't want to send invalid state to the client and have them display a bogus
		// completion percentage.  If we detect an invalid state, just send the last
		// known good values.
		lastRatio := float64(lastProgress) / float64(lastExpected)
		if status.Progress > 0 &&
			status.Expected > 0 &&
			status.Progress <= status.Expected &&
			lastRatio < (float64(status.Progress)/float64(status.Expected)) {
			lastProgress = status.Progress
			lastExpected = status.Expected
		} else {
			status.Progress = lastProgress
			status.Expected = lastExpected
		}

		tag := reflect.TypeOf(&extint.GatewayWrapper_Event{}).String()
		msg := extint.GatewayWrapper{
			OneofMessageType: &extint.GatewayWrapper_Event{
				// TODO: Convert all events into proto events
				Event: &extint.Event{
					EventType: &extint.Event_CheckUpdateStatusResponse{
						CheckUpdateStatusResponse: status,
					},
				},
			},
		}

		if logVerbose {
			log.Printf("%s, err = %s, exit = %d, phase = %s, ver = %s, prog = %d, exp = %d",
				status.UpdateStatus,
				status.Error,
				status.ExitCode,
				status.UpdatePhase,
				status.UpdateVersion,
				status.Progress,
				status.Expected)
		}

		engineProtoManager.SendToListeners(tag, msg)

		// If we have an ExitCode and we have sent at least 2 updates, we can stop
		// sending status updates as we will just be repeating ourselves.  We could
		// probably stop after a single update, but the smartphone app doesn't seem
		// to display the very first status change it receives.
		if status.ExitCode != otaExitCodeNotAvailable && updateStarted {
			break
		}
		updateStarted = true
		time.Sleep(2000 * time.Millisecond)
	}
}

// StartUpdateEngine restarts the update-engine process and starts a stream of status messages to the app.
func (service *rpcService) StartUpdateEngine(
	ctx context.Context, in *extint.CheckUpdateStatusRequest) (*extint.CheckUpdateStatusResponse, error) {

	retval := &extint.CheckUpdateStatusResponse{
		Status: &extint.ResponseStatus{
			Code: extint.ResponseStatus_RESPONSE_RECEIVED,
		},
	}

	status, _ := service.GetUpdateStatus()

	// Unless we are ready to reboot into a new OS version, we need to make sure that
	// /anki/bin/update-engine is running.  By the way, /anki/bin/update-engine will NOT
	// be considered active if it is doing its 'sleeping' phase.
	restartUpdateEngine := false
	if status.UpdateStatus != extint.CheckUpdateStatusResponse_READY_TO_REBOOT_INTO_NEW_OS_VERSION {
		err := exec.Command(
			"/bin/systemctl",
			"is-active",
			"--quiet",
			"update-engine.service").Run()
		restartUpdateEngine = err != nil
	}

	if restartUpdateEngine {
		err := exec.Command(
			"/usr/bin/sudo",
			"-n",
			"/bin/systemctl",
			"stop",
			"update-engine.service").Run()
		if err != nil {
			log.Errorf("Update attempt failed on `systemctl stop update-engine`: %s\n", err)
			retval.Status.Code = extint.ResponseStatus_ERROR_UPDATE_IN_PROGRESS
			return retval, err
		}

		err = exec.Command(
			"/usr/bin/sudo",
			"-n",
			"/bin/systemctl",
			"restart",
			"update-engine-oneshot").Run()
		if err != nil {
			log.Errorf("Update attempt failed on `systemctl restart update-engine-oneshot`: %s\n", err)
			retval.Status.Code = extint.ResponseStatus_ERROR_UPDATE_IN_PROGRESS
			return retval, err
		}
	}

	// Reset the lastProgress and lastExpected globals that are used in UpdateStatusStream, so that
	// we are sure to send reasonable starting progress values to the smartphone app
	lastProgress = 0
	lastExpected = 1
	go service.UpdateStatusStream()

	return retval, nil
}

// CheckUpdateStatus tells if the robot is ready to reboot and update.
func (service *rpcService) CheckUpdateStatus(
	ctx context.Context, in *extint.CheckUpdateStatusRequest) (*extint.CheckUpdateStatusResponse, error) {

	return service.GetUpdateStatus()
}

// UpdateAndRestart reboots the robot when an update is available.
// This will apply the update when the robot starts up.
func (service *rpcService) UpdateAndRestart(ctx context.Context, in *extint.UpdateAndRestartRequest) (*extint.UpdateAndRestartResponse, error) {
	if _, err := os.Stat("/run/update-engine/done"); err == nil {
		go func() {
			<-time.After(5 * time.Second)
			err := exec.Command("/usr/bin/sudo", "/sbin/reboot").Run()
			if err != nil {
				log.Errorf("Reboot attempt failed: %s\n", err)
			}
		}()
		return &extint.UpdateAndRestartResponse{
			Status: &extint.ResponseStatus{
				Code: extint.ResponseStatus_OK,
			},
		}, nil
	}
	return &extint.UpdateAndRestartResponse{
		Status: &extint.ResponseStatus{
			Code: extint.ResponseStatus_NOT_FOUND,
		},
	}, nil
}

// UploadDebugLogs will upload debug logs to S3, and return a url to the caller.
func (service *rpcService) UploadDebugLogs(ctx context.Context, in *extint.UploadDebugLogsRequest) (*extint.UploadDebugLogsResponse, error) {
	if !debugLogLimiter.Allow() {
		return nil, grpc.Errorf(codes.ResourceExhausted, "Maximum upload rate exceeded. Please wait and try again later.")
	}

	/* disabling so we can build the gateway -bd

	url, err := loguploader.UploadDebugLogs()
	if err != nil {
		log.Println("MessageHandler.UploadDebugLogs.Error: " + err.Error())
		return nil, grpc.Errorf(codes.Internal, err.Error())
	}
	response := &extint.UploadDebugLogsResponse{
		Status: &extint.ResponseStatus{
			Code: extint.ResponseStatus_OK,
		},
		Url: url,
	}
	return response, nil
	*/

	response := &extint.UploadDebugLogsResponse{
		Status: &extint.ResponseStatus{
			Code: extint.ResponseStatus_OK,
		},
	}
	return response, nil
}

var lastResult *extint.CheckCloudResponse

// CheckCloudConnection is used to verify Vector's connection to the Anki Cloud
// Its main use is to be called by the app during setup, but is fine for use by the outside world.
func (service *rpcService) CheckCloudConnection(ctx context.Context, in *extint.CheckCloudRequest) (*extint.CheckCloudResponse, error) {
	if !cloudCheckLimiter.Allow() {
		if lastResult == nil {
			lastResult = &extint.CheckCloudResponse{
				Status: &extint.ResponseStatus{
					Code: extint.ResponseStatus_UNKNOWN,
				},
				Code: extint.CheckCloudResponse_UNKNOWN,
			}
		}
		return lastResult, nil
		// TODO: change this back to a resource exhausted error after app properly handles the error
		// return nil, grpc.Errorf(codes.ResourceExhausted, "Maximum check rate exceeded. Please wait and try again later.")
	}

	f, responseChan := engineProtoManager.CreateChannel(&extint.GatewayWrapper_CheckCloudResponse{}, 1)
	defer f()

	_, _, err := engineProtoManager.Write(&extint.GatewayWrapper{
		OneofMessageType: &extint.GatewayWrapper_CheckCloudRequest{
			CheckCloudRequest: in,
		},
	})
	if err != nil {
		return nil, err
	}
	payload, ok := <-responseChan
	if !ok {
		return nil, grpc.Errorf(codes.Internal, "Failed to retrieve message")
	}
	cloudResponse := payload.GetCheckCloudResponse()
	cloudResponse.Status = &extint.ResponseStatus{
		Code: extint.ResponseStatus_RESPONSE_RECEIVED,
	}
	lastResult = cloudResponse
	return cloudResponse, nil
}

func (service *rpcService) DeleteCustomObjects(ctx context.Context, in *extint.DeleteCustomObjectsRequest) (*extint.DeleteCustomObjectsResponse, error) {
	var responseMessageType gw_clad.MessageRobotToExternalTag
	var cladMsg *gw_clad.MessageExternalToRobot

	switch in.Mode {
	case extint.CustomObjectDeletionMode_DELETION_MASK_ARCHETYPES:
		responseMessageType = gw_clad.MessageRobotToExternalTag_RobotDeletedCustomMarkerObjects
		cladMsg = gw_clad.NewMessageExternalToRobotWithUndefineAllCustomMarkerObjects(
			&gw_clad.UndefineAllCustomMarkerObjects{})
		break
	case extint.CustomObjectDeletionMode_DELETION_MASK_FIXED_CUSTOM_OBJECTS:
		responseMessageType = gw_clad.MessageRobotToExternalTag_RobotDeletedFixedCustomObjects
		cladMsg = gw_clad.NewMessageExternalToRobotWithDeleteFixedCustomObjects(
			&gw_clad.DeleteFixedCustomObjects{})
		break
	case extint.CustomObjectDeletionMode_DELETION_MASK_CUSTOM_MARKER_OBJECTS:
		responseMessageType = gw_clad.MessageRobotToExternalTag_RobotDeletedCustomMarkerObjects
		cladMsg = gw_clad.NewMessageExternalToRobotWithDeleteCustomMarkerObjects(
			&gw_clad.DeleteCustomMarkerObjects{})
		break
	}

	f, responseChan := engineCladManager.CreateChannel(responseMessageType, 1)
	defer f()

	_, err := engineCladManager.Write(cladMsg)

	if err != nil {
		return nil, err
	}

	_, ok := <-responseChan
	if !ok {
		return nil, grpc.Errorf(codes.Internal, "Failed to retrieve message")
	}

	return &extint.DeleteCustomObjectsResponse{
		Status: &extint.ResponseStatus{
			Code: extint.ResponseStatus_RESPONSE_RECEIVED,
		},
	}, nil
}

func (service *rpcService) CreateFixedCustomObject(ctx context.Context, in *extint.CreateFixedCustomObjectRequest) (*extint.CreateFixedCustomObjectResponse, error) {
	f, responseChan := engineCladManager.CreateChannel(gw_clad.MessageRobotToExternalTag_CreatedFixedCustomObject, 1)
	defer f()

	cladData := ProtoCreateFixedCustomObjectToClad(in)

	_, err := engineCladManager.Write(cladData)
	if err != nil {
		return nil, err
	}

	chanResponse, ok := <-responseChan
	if !ok {
		return nil, grpc.Errorf(codes.Internal, "Failed to retrieve message")
	}
	response := chanResponse.GetCreatedFixedCustomObject()

	return &extint.CreateFixedCustomObjectResponse{
		Status: &extint.ResponseStatus{
			Code: extint.ResponseStatus_RESPONSE_RECEIVED,
		},
		ObjectId: response.ObjectID,
	}, nil
}

func (service *rpcService) DefineCustomObject(ctx context.Context, in *extint.DefineCustomObjectRequest) (*extint.DefineCustomObjectResponse, error) {
	var cladMsg *gw_clad.MessageExternalToRobot

	f, responseChan := engineCladManager.CreateChannel(gw_clad.MessageRobotToExternalTag_DefinedCustomObject, 1)
	defer f()

	switch x := in.CustomObjectDefinition.(type) {
	case *extint.DefineCustomObjectRequest_CustomBox:
		cladMsg = ProtoDefineCustomBoxToClad(in, in.GetCustomBox())
		break
	case *extint.DefineCustomObjectRequest_CustomCube:
		cladMsg = ProtoDefineCustomCubeToClad(in, in.GetCustomCube())
		break
	case *extint.DefineCustomObjectRequest_CustomWall:
		cladMsg = ProtoDefineCustomWallToClad(in, in.GetCustomWall())
		break
	default:
		return nil, grpc.Errorf(codes.InvalidArgument, "DefineCustomObjectRequest has unexpected type %T", x)
	}

	_, err := engineCladManager.Write(cladMsg)
	if err != nil {
		return nil, err
	}

	chanResponse, ok := <-responseChan
	if !ok {
		return nil, grpc.Errorf(codes.Internal, "Failed to retrieve message")
	}
	response := chanResponse.GetDefinedCustomObject()

	return &extint.DefineCustomObjectResponse{
		Status: &extint.ResponseStatus{
			Code: extint.ResponseStatus_RESPONSE_RECEIVED,
		},
		Success: response.Success,
	}, nil
}

// FeatureFlag is used to check what features are enabled on the robot
func (service *rpcService) GetFeatureFlag(ctx context.Context, in *extint.FeatureFlagRequest) (*extint.FeatureFlagResponse, error) {
	f, responseChan := engineProtoManager.CreateChannel(&extint.GatewayWrapper_FeatureFlagResponse{}, 1)
	defer f()

	_, _, err := engineProtoManager.Write(&extint.GatewayWrapper{
		OneofMessageType: &extint.GatewayWrapper_FeatureFlagRequest{
			FeatureFlagRequest: in,
		},
	})
	if err != nil {
		return nil, err
	}
	payload, ok := <-responseChan
	if !ok {
		return nil, grpc.Errorf(codes.Internal, "Failed to retrieve message")
	}
	response := payload.GetFeatureFlagResponse()
	response.Status = &extint.ResponseStatus{
		Code: extint.ResponseStatus_RESPONSE_RECEIVED,
	}
	return response, nil
}

// FeatureFlagList is used to check what features are enabled on the robot
func (service *rpcService) GetFeatureFlagList(ctx context.Context, in *extint.FeatureFlagListRequest) (*extint.FeatureFlagListResponse, error) {
	f, responseChan := engineProtoManager.CreateChannel(&extint.GatewayWrapper_FeatureFlagListResponse{}, 1)
	defer f()

	_, _, err := engineProtoManager.Write(&extint.GatewayWrapper{
		OneofMessageType: &extint.GatewayWrapper_FeatureFlagListRequest{
			FeatureFlagListRequest: in,
		},
	})
	if err != nil {
		return nil, err
	}
	payload, ok := <-responseChan
	if !ok {
		return nil, grpc.Errorf(codes.Internal, "Failed to retrieve message")
	}
	response := payload.GetFeatureFlagListResponse()
	response.Status = &extint.ResponseStatus{
		Code: extint.ResponseStatus_RESPONSE_RECEIVED,
	}
	return response, nil
}

// AlexaAuthState is used to check the alexa authorization state
func (service *rpcService) GetAlexaAuthState(ctx context.Context, in *extint.AlexaAuthStateRequest) (*extint.AlexaAuthStateResponse, error) {
	f, responseChan := engineProtoManager.CreateChannel(&extint.GatewayWrapper_AlexaAuthStateResponse{}, 1)
	defer f()

	_, _, err := engineProtoManager.Write(&extint.GatewayWrapper{
		OneofMessageType: &extint.GatewayWrapper_AlexaAuthStateRequest{
			AlexaAuthStateRequest: in,
		},
	})
	if err != nil {
		return nil, err
	}
	payload, ok := <-responseChan
	if !ok {
		return nil, grpc.Errorf(codes.Internal, "Failed to retrieve message")
	}
	response := payload.GetAlexaAuthStateResponse()
	response.Status = &extint.ResponseStatus{
		Code: extint.ResponseStatus_RESPONSE_RECEIVED,
	}
	return response, nil
}

// AlexaOptIn is used to check the alexa authorization state
func (service *rpcService) AlexaOptIn(ctx context.Context, in *extint.AlexaOptInRequest) (*extint.AlexaOptInResponse, error) {
	_, _, err := engineProtoManager.Write(&extint.GatewayWrapper{
		OneofMessageType: &extint.GatewayWrapper_AlexaOptInRequest{
			AlexaOptInRequest: in,
		},
	})
	if err != nil {
		return nil, err
	}
	return &extint.AlexaOptInResponse{
		Status: &extint.ResponseStatus{
			Code: extint.ResponseStatus_REQUEST_PROCESSING,
		},
	}, nil
}

func SetNavMapBroadcastFrequency(frequency float32) error {
	log.Println("Setting NavMapBroadcastFrequency to (", frequency, ") seconds")

	cladMsg := gw_clad.NewMessageExternalToRobotWithSetMemoryMapBroadcastFrequencySec(&gw_clad.SetMemoryMapBroadcastFrequency_sec{
		Frequency: frequency,
	})
	_, err := engineCladManager.Write(cladMsg)

	return err
}

func (service *rpcService) NavMapFeed(in *extint.NavMapFeedRequest, stream extint.ExternalInterface_NavMapFeedServer) error {

	// Enable nav map stream
	err := SetNavMapBroadcastFrequency(in.Frequency)
	if err != nil {
		return err
	}

	// Disable nav map stream when the RPC exits
	defer SetNavMapBroadcastFrequency(-1.0)

	f1, memoryMapMessageBegin := engineCladManager.CreateChannel(gw_clad.MessageRobotToExternalTag_MemoryMapMessageBegin, 1)
	defer f1()

	// Every frame the engine can send up to: (Anki::Comms::MsgPacket::MAX_SIZE-3)/sizeof(QuadInfoVector::value_type) quads.
	// 50 feels like a reasonable educated guess.
	f2, memoryMapMessageData := engineCladManager.CreateChannel(gw_clad.MessageRobotToExternalTag_MemoryMapMessage, 50)
	defer f2()

	f3, memoryMapMessageEnd := engineCladManager.CreateChannel(gw_clad.MessageRobotToExternalTag_MemoryMapMessageEnd, 1)
	defer f3()

	var pendingMap *extint.NavMapFeedResponse = nil

	for {
		select {
		case chanResponse, ok := <-memoryMapMessageBegin:
			if !ok {
				return grpc.Errorf(codes.Internal, "Failed to retrieve message")
			}
			if pendingMap != nil {
				log.Println("MessageHandler.NavMapFeed.Error: MemoryMapBegin received from engine while still processing a pending memory map; discarding pending map.")
			}

			response := chanResponse.GetMemoryMapMessageBegin()
			pendingMap = &extint.NavMapFeedResponse{
				OriginId:  response.OriginId,
				MapInfo:   CladMemoryMapBeginToProtoNavMapInfo(response),
				QuadInfos: []*extint.NavMapQuadInfo{},
			}

		case chanResponse, ok := <-memoryMapMessageData:
			if !ok {
				return grpc.Errorf(codes.Internal, "Failed to retrieve message")
			}
			if pendingMap == nil {
				log.Println("MessageHandler.NavMapFeed.Error: MemoryMapData received from engine with no pending content to add to.")
			} else {
				response := chanResponse.GetMemoryMapMessage()
				for i := 0; i < len(response.QuadInfos); i++ {
					newQuad := CladMemoryMapQuadInfoToProto(&response.QuadInfos[i])
					pendingMap.QuadInfos = append(pendingMap.QuadInfos, newQuad)
				}
			}

		case _, ok := <-memoryMapMessageEnd:
			if !ok {
				return grpc.Errorf(codes.Internal, "Failed to retrieve message")
			}

			if pendingMap == nil {
				log.Println("MessageHandler.NavMapFeed.Error: MemoryMapEnd received from engine with no pending content to send.")
			} else if err := stream.Send(pendingMap); err != nil {
				return err
			} else if err = stream.Context().Err(); err != nil {
				// This is the case where the user disconnects the stream
				// We should still return the err in case the user doesn't think they disconnected
				return err
			}

			pendingMap = nil
		}
	}

	errMsg := "NavMemoryMap engine stream died unexpectedly"
	log.Errorln(errMsg)
	return grpc.Errorf(codes.Internal, errMsg)
}

func newServer() *rpcService {
	return new(rpcService)
}

// Set Eye Color (SDK only)
// TODO Set eye color back to Settings value in internal code when SDK program ends or loses behavior control
// (e.g., in go code or when SDK behavior deactivates)
func (service *rpcService) SetEyeColor(ctx context.Context, in *extint.SetEyeColorRequest) (*extint.SetEyeColorResponse, error) {
	_, _, err := engineProtoManager.Write(&extint.GatewayWrapper{
		OneofMessageType: &extint.GatewayWrapper_SetEyeColorRequest{
			SetEyeColorRequest: in,
		},
	})
	if err != nil {
		return nil, err
	}
	return &extint.SetEyeColorResponse{
		Status: &extint.ResponseStatus{
			Code: extint.ResponseStatus_REQUEST_PROCESSING,
		},
	}, nil
}

// Get Camera Configuration
func (service *rpcService) GetCameraConfig(ctx context.Context, in *extint.CameraConfigRequest) (*extint.CameraConfigResponse, error) {
	f, responseChan := engineProtoManager.CreateChannel(&extint.GatewayWrapper_CameraConfigResponse{}, 1)
	defer f()

	_, _, err := engineProtoManager.Write(&extint.GatewayWrapper{
		OneofMessageType: &extint.GatewayWrapper_CameraConfigRequest{
			CameraConfigRequest: in,
		},
	})
	if err != nil {
		return nil, err
	}

	payload, ok := <-responseChan
	if !ok {
		return nil, grpc.Errorf(codes.Internal, "Failed to retrieve message")
	}
	response := payload.GetCameraConfigResponse()
	return response, nil
}

// Set Camera Settings
func (service *rpcService) SetCameraSettings(ctx context.Context, in *extint.SetCameraSettingsRequest) (*extint.SetCameraSettingsResponse, error) {
	f, responseChan := engineProtoManager.CreateChannel(&extint.GatewayWrapper_SetCameraSettingsResponse{}, 1)
	defer f()

	_, _, err := engineProtoManager.Write(&extint.GatewayWrapper{
		OneofMessageType: &extint.GatewayWrapper_SetCameraSettingsRequest{
			SetCameraSettingsRequest: in,
		},
	})
	if err != nil {
		return nil, err
	}

	payload, ok := <-responseChan
	if !ok {
		return nil, grpc.Errorf(codes.Internal, "Failed to retrieve message")
	}
	response := payload.GetSetCameraSettingsResponse()
	response.Status = &extint.ResponseStatus{
		Code: extint.ResponseStatus_RESPONSE_RECEIVED,
	}
	return response, nil
}

func (service *rpcService) ExternalAudioStreamRequestToGatewayWrapper(request *extint.ExternalAudioStreamRequest) (*extint.GatewayWrapper, error) {
	msg := &extint.GatewayWrapper{}
	switch x := request.AudioRequestType.(type) {
	case *extint.ExternalAudioStreamRequest_AudioStreamPrepare:
		msg.OneofMessageType = &extint.GatewayWrapper_ExternalAudioStreamPrepare{
			ExternalAudioStreamPrepare: request.GetAudioStreamPrepare(),
		}
	case *extint.ExternalAudioStreamRequest_AudioStreamChunk:
		msg.OneofMessageType = &extint.GatewayWrapper_ExternalAudioStreamChunk{
			ExternalAudioStreamChunk: request.GetAudioStreamChunk(),
		}
	case *extint.ExternalAudioStreamRequest_AudioStreamComplete:
		msg.OneofMessageType = &extint.GatewayWrapper_ExternalAudioStreamComplete{
			ExternalAudioStreamComplete: request.GetAudioStreamComplete(),
		}
	case *extint.ExternalAudioStreamRequest_AudioStreamCancel:
		msg.OneofMessageType = &extint.GatewayWrapper_ExternalAudioStreamCancel{
			ExternalAudioStreamCancel: request.GetAudioStreamCancel(),
		}
	default:
		return nil, grpc.Errorf(codes.InvalidArgument, "ExternalAudioStreamRequest.AudioStreamControlRequest has unexpected type %T", x)
	}

	return msg, nil
}

func (service *rpcService) ExternalAudioStreamRequestHandler(in extint.ExternalInterface_ExternalAudioStreamPlaybackServer, done chan struct{}) {
	defer close(done)
	for {
		request, err := in.Recv()
		if err != nil {
			log.Printf("AudioStreamRequestHandler.close: %s\n", err.Error())
			return
		}
		log.Println("External AudioStream playback incoming request") //not printing message, lengthy sample data too busy for logs

		msg, err := service.ExternalAudioStreamRequestToGatewayWrapper(request)
		if err != nil {
			log.Println(err)
			return
		}

		numCommandsSentFromSDK++
		_, _, err = engineProtoManager.Write(msg)
		if err != nil {
			log.Printf("Could not write GatewayWrapper_AudioStreamRequest\n")
		}
	}
}

func (service *rpcService) ExternalAudioStreamResponseHandler(out extint.ExternalInterface_ExternalAudioStreamPlaybackServer, responses chan extint.GatewayWrapper, done chan struct{}) error {
	for {
		select {
		case <-done:
			return nil
		case response, ok := <-responses:
			if !ok {
				return grpc.Errorf(codes.Internal, "Failed to retrieve message")
			}
			msg := response.GetExternalAudioStreamResponse()
			if err := out.Send(msg); err != nil {
				log.Println("Closing AudioStream (on send):", err)
				return err
			} else if err = out.Context().Err(); err != nil {
				// This is the case where the user disconnects the stream
				// We should still return the err in case the user doesn't think they disconnected
				log.Println("Closing AudioStream stream:", err)
				return err
			}
		}
	}
	return nil
}

// Stream audio to the robot, stream audio playback status back
func (service *rpcService) ExternalAudioStreamPlayback(bidirectionalStream extint.ExternalInterface_ExternalAudioStreamPlaybackServer) error {
	audioStartTime := time.Now()

	log.Println("sdk.audiostream_started")

	defer func() {
		sdkElapsedSeconds := time.Since(audioStartTime)
		log.Printf("sdk.audiostream_ended %s\n", sdkElapsedSeconds.String())
	}()

	done := make(chan struct{})

	f, audioStreamStatus := engineProtoManager.CreateChannel(&extint.GatewayWrapper_ExternalAudioStreamResponse{}, 1)
	defer f()

	go service.ExternalAudioStreamRequestHandler(bidirectionalStream, done)
	return service.ExternalAudioStreamResponseHandler(bidirectionalStream, audioStreamStatus, done)
}
