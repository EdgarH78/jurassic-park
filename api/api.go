package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/EdgarH78/jurassic-park/models"
	"github.com/gin-gonic/gin"
)

const baseUrl = "jurassicpark/v1"

type parkManager interface {
	AddCage(cage models.Cage) error
	GetCage(cageLabel string) (*models.Cage, error)
	GetCages() ([]models.Cage, error)
	AddDinosaur(dinosaur models.Dinosaur) error
	GetDinosaurs() ([]models.Dinosaur, error)
	GetDinosaur(name string) (*models.Dinosaur, error)
	AddDinosaurToCage(dinosaurName, targetCage string) error
	GetDinosaursInCage(cageLabel string) ([]models.Dinosaur, error)
	UpdateCagePowerStatus(cageLabel string, powerOn bool) error
}

type API struct {
	engine      *gin.Engine
	parkManager parkManager
}

func NewAPI(parkManager parkManager, engine *gin.Engine) *API {
	api := &API{
		parkManager: parkManager,
		engine:      engine,
	}

	api.registerHandlers()
	return api
}

func (api *API) Run() {
	api.engine.Run(":8080")
}

func (api *API) registerHandlers() {
	api.engine.POST(baseUrl+"/cages", api.CreateCage)
	api.engine.GET(baseUrl+"/cages", api.GetCages)
	api.engine.GET(baseUrl+"/cages/:cageLabel", api.GetCage)
	api.engine.PATCH(baseUrl+"/cages/:cageLabel", api.UpdateCagePowerStatus)
	api.engine.GET(baseUrl+"/cages/:cageLabel/dinosaurs", api.GetDinosaursInCage)
	api.engine.POST(baseUrl+"/cages/:cageLabel/dinosaurs", api.AddDinosaurToCage)
	api.engine.POST(baseUrl+"/dinosaurs", api.AddDinosaur)
	api.engine.GET(baseUrl+"/dinosaurs", api.GetDinosaurs)
	api.engine.GET(baseUrl+"/dinosaurs/:name", api.GetDinosaur)
}

func (api *API) CreateCage(c *gin.Context) {

	var cage models.Cage
	err := json.NewDecoder(c.Request.Body).Decode(&cage)
	if err != nil {
		c.String(http.StatusUnprocessableEntity, "Request body is in the incorrect format")
		return
	}
	err = api.parkManager.AddCage(cage)
	if err != nil {
		c.String(http.StatusInternalServerError, "An error occured while adding the cage")
	} else {
		c.String(http.StatusCreated, "cage created")
	}
}

func (api *API) GetCages(c *gin.Context) {
	cages, err := api.parkManager.GetCages()
	if err != nil {
		c.String(http.StatusInternalServerError, "unexpected error")
		return
	}
	c.JSON(http.StatusOK, cages)
}

func (api *API) GetCage(c *gin.Context) {
	cageLabel := c.Param("cageLabel")
	cage, err := api.parkManager.GetCage(cageLabel)
	if err != nil {
		if errors.Is(err, models.EntityNotFound) {
			c.String(http.StatusNotFound, fmt.Sprintf("cage with label %s not found", cageLabel))
		} else {
			c.String(http.StatusInternalServerError, "unexpected error")
		}
		return
	}

	c.JSON(http.StatusOK, cage)
}

func (api *API) UpdateCagePowerStatus(c *gin.Context) {
	cageLabel := c.Param("cageLabel")
	var updatePowerStatusRequest models.UpdateCagePowerStatusRequest
	err := json.NewDecoder(c.Request.Body).Decode(&updatePowerStatusRequest)
	if err != nil {
		c.String(http.StatusUnprocessableEntity, "Request body is in the incorrect format")
		return
	}

	err = api.parkManager.UpdateCagePowerStatus(cageLabel, updatePowerStatusRequest.HasPower)
	if err != nil {
		if errors.Is(err, models.IncompatibleCagePowerState) {
			c.String(http.StatusConflict, "the cage has dinosaurs in it and cannot be powered off")
		} else if errors.Is(err, models.EntityNotFound) {
			c.String(http.StatusNotFound, "could not find cage")
		} else {
			c.String(http.StatusInternalServerError, "unexpected error")
		}
		return
	}
	c.String(http.StatusOK, "power status updated")
}

func (api *API) GetDinosaursInCage(c *gin.Context) {
	cageLabel := c.Param("cageLabel")
	dinosaurs, err := api.parkManager.GetDinosaursInCage(cageLabel)
	if err != nil {
		if errors.Is(err, models.EntityNotFound) {
			c.String(http.StatusNotFound, fmt.Sprintf("the cage %s was not found", cageLabel))
		} else {
			c.String(http.StatusInternalServerError, "unexpected error")
		}
		return
	}
	c.JSON(http.StatusOK, dinosaurs)
}

func (api *API) AddDinosaurToCage(c *gin.Context) {
	var addDinosaurRequest models.AddDinosaurToCageRequest
	err := json.NewDecoder(c.Request.Body).Decode(&addDinosaurRequest)
	if err != nil {
		c.String(http.StatusUnprocessableEntity, "Request body is in the incorrect format")
		return
	}
	targetCage := c.Param("cageLabel")
	err = api.parkManager.AddDinosaurToCage(addDinosaurRequest.Name, targetCage)
	if err != nil {
		if errors.Is(err, models.CageCapacityExceeded) {
			c.String(http.StatusConflict, "the cage is at capacity")
		} else if errors.Is(err, models.IncompatibleCagePowerState) {
			c.String(http.StatusConflict, "the cage is unavailable at this time, because it does not have power")
		} else if errors.Is(err, models.IncompatibleSpecies) {
			c.String(http.StatusConflict, "the cage already contains species of dinosaur that are incompatible with this dinosaur's specie")
		} else if errors.Is(err, models.EntityNotFound) {
			c.String(http.StatusNotFound, "could not find either the cage or dinosaur")
		} else {
			c.String(http.StatusInternalServerError, "unexpected error")
		}
		return
	}
	c.String(http.StatusCreated, "dinosaur added")
}

func (api *API) AddDinosaur(c *gin.Context) {
	var dinosaur models.Dinosaur
	err := json.NewDecoder(c.Request.Body).Decode(&dinosaur)
	if err != nil {
		c.String(http.StatusUnprocessableEntity, "Request body is in the incorrect format")
		return
	}

	err = api.parkManager.AddDinosaur(dinosaur)
	if err != nil {
		if errors.Is(err, models.InvalidDinosaurSpecies) {
			c.String(http.StatusConflict, fmt.Sprintf("The species %s is not a valid dinosaur species", dinosaur.Species))
		} else if errors.Is(err, models.EntityAlreadyExists) {
			c.String(http.StatusConflict, fmt.Sprintf("There is already a dinosaur with the name %s", dinosaur.Name))
		}
		return
	}
	c.String(http.StatusCreated, "dinosaur created")
}

func (api *API) GetDinosaurs(c *gin.Context) {
	dinosaurs, err := api.parkManager.GetDinosaurs()
	if err != nil {
		c.String(http.StatusInternalServerError, "unexpected error")
		return
	}
	c.JSON(http.StatusOK, dinosaurs)
}

func (api *API) GetDinosaur(c *gin.Context) {
	dinosaurName := c.Param("name")
	dinosaur, err := api.parkManager.GetDinosaur(dinosaurName)
	if err != nil {
		if errors.Is(err, models.EntityNotFound) {
			c.String(http.StatusNotFound, fmt.Sprintf("dinosaur with name %s not found", dinosaurName))
		} else {
			c.String(http.StatusInternalServerError, "unexpected error")
		}
		return
	}
	c.JSON(http.StatusOK, dinosaur)
}
