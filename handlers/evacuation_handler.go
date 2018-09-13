package handlers

import (
	"net/http"

	"code.cloudfoundry.org/bbs/controllers"
	"code.cloudfoundry.org/bbs/models"
	"code.cloudfoundry.org/lager"
	"github.com/gogo/protobuf/proto"
)

//go:generate counterfeiter -o fake_controllers/fake_evacuation_controller.go . EvacuationController
type EvacuationController interface {
	RemoveEvacuatingActualLRP(logger lager.Logger, w http.ResponseWriter, req *http.Request) error
	EvacuateClaimedActualLRP(logger lager.Logger, w http.ResponseWriter, req *http.Request) (error, bool)
	EvacuateCrashedActualLRP(logger lager.Logger, w http.ResponseWriter, req *http.Request) (error, bool)
	EvacuateRunningActualLRP(logger lager.Logger, w http.ResponseWriter, req *http.Request) (error, bool)
	EvacuateStoppedActualLRP(logger lager.Logger, w http.ResponseWriter, req *http.Request) (error, bool)
}

type EvacuationHandler struct {
	controller *controllers.EvacuationController
	exitChan   chan<- struct{}
}

func NewEvacuationHandler(
	controller *controllers.EvacuationController,
	exitChan chan<- struct{},
) *EvacuationHandler {
	return &EvacuationHandler{
		controller: controller,
		exitChan:   exitChan,
	}
}

type MessageValidator interface {
	proto.Message
	Validate() error
	Unmarshal(data []byte) error
}

func (h *EvacuationHandler) RemoveEvacuatingActualLRP(logger lager.Logger, w http.ResponseWriter, req *http.Request) {
	var err error
	logger = logger.Session("remove-evacuating-actual-lrp")
	logger.Info("started")
	defer logger.Info("completed")

	request := &models.RemoveEvacuatingActualLRPRequest{}
	response := &models.RemoveEvacuatingActualLRPResponse{}

	defer func() { exitIfUnrecoverable(logger, h.exitChan, response.Error) }()
	defer writeResponse(w, response)

	err = parseRequest(logger, req, request)
	if err != nil {
		response.Error = models.ConvertError(err)
		return
	}

	err = h.controller.RemoveEvacuatingActualLRP(logger, request.ActualLrpKey, request.ActualLrpInstanceKey)
	response.Error = models.ConvertError(err)
}

func (h *EvacuationHandler) EvacuateClaimedActualLRP(logger lager.Logger, w http.ResponseWriter, req *http.Request) {
	logger = logger.Session("evacuate-claimed-actual-lrp")
	logger.Info("started")
	defer logger.Info("completed")

	request := &models.EvacuateClaimedActualLRPRequest{}
	response := &models.EvacuationResponse{}
	defer func() { exitIfUnrecoverable(logger, h.exitChan, response.Error) }()
	defer writeResponse(w, response)

	err := parseRequest(logger, req, request)
	if err != nil {
		logger.Error("failed-parsing-request", err)
		response.Error = models.ConvertError(err)
		response.KeepContainer = true
		return
	}

	err, keepContainer := h.controller.EvacuateClaimedActualLRP(logger, request.ActualLrpKey, request.ActualLrpInstanceKey)
	response.Error = models.ConvertError(err)
	response.KeepContainer = keepContainer
}

func (h *EvacuationHandler) EvacuateCrashedActualLRP(logger lager.Logger, w http.ResponseWriter, req *http.Request) {
	logger = logger.Session("evacuate-crashed-actual-lrp")
	logger.Info("started")
	defer logger.Info("completed")

	request := &models.EvacuateCrashedActualLRPRequest{}
	response := &models.EvacuationResponse{}
	defer func() { exitIfUnrecoverable(logger, h.exitChan, response.Error) }()
	defer writeResponse(w, response)

	err := parseRequest(logger, req, request)
	if err != nil {
		logger.Error("failed-parsing-request", err)
		response.Error = models.ConvertError(err)
		return
	}

	err = h.controller.EvacuateCrashedActualLRP(logger, request.ActualLrpKey, request.ActualLrpInstanceKey, request.ErrorMessage)
	response.Error = models.ConvertError(err)
}

func (h *EvacuationHandler) EvacuateRunningActualLRP(logger lager.Logger, w http.ResponseWriter, req *http.Request) {
	logger = logger.Session("evacuate-running-actual-lrp")
	logger.Info("starting")
	defer logger.Info("completed")

	response := &models.EvacuationResponse{}
	response.KeepContainer = true
	defer func() { exitIfUnrecoverable(logger, h.exitChan, response.Error) }()
	defer writeResponse(w, response)

	request := &models.EvacuateRunningActualLRPRequest{}
	err := parseRequest(logger, req, request)
	if err != nil {
		response.Error = models.ConvertError(err)
		return
	}

	err, keepContainer := h.controller.EvacuateRunningActualLRP(logger, request.ActualLrpKey, request.ActualLrpInstanceKey, request.ActualLrpNetInfo)
	response.Error = models.ConvertError(err)
	response.KeepContainer = keepContainer
}

func (h *EvacuationHandler) EvacuateStoppedActualLRP(logger lager.Logger, w http.ResponseWriter, req *http.Request) {
	logger = logger.Session("evacuate-stopped-actual-lrp")

	request := &models.EvacuateStoppedActualLRPRequest{}
	response := &models.EvacuationResponse{}

	defer func() { exitIfUnrecoverable(logger, h.exitChan, response.Error) }()
	defer writeResponse(w, response)

	err := parseRequest(logger, req, request)
	if err != nil {
		logger.Error("failed-to-parse-request", err)
		response.Error = models.ConvertError(err)
		return
	}

	err = h.controller.EvacuateStoppedActualLRP(logger, request.ActualLrpKey, request.ActualLrpInstanceKey)
	response.Error = models.ConvertError(err)
}
