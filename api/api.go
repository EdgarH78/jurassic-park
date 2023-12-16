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
	GetCages(filter models.CageFilter) ([]models.Cage, error)
	AddDinosaur(dinosaur models.Dinosaur) error
	GetDinosaurs(filter models.DinosaurFilter) ([]models.Dinosaur, error)
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
		c.JSON(http.StatusUnprocessableEntity, models.ErrorResponse{
			ErrorMessage: "Request body is in the incorrect format",
		})
		return
	}
	err = api.parkManager.AddCage(cage)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			ErrorMessage: "An error occured while adding the cage",
		})
	} else {
		c.JSON(http.StatusCreated, cage)
	}
}

func (api *API) GetCages(c *gin.Context) {
	filter := models.CageFilter{}
	if c.Query("hasPower") != "" {
		hasPower := c.Query("hasPower") == "true"
		filter.HasPower = &hasPower
	}
	cages, err := api.parkManager.GetCages(filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			ErrorMessage: "unexpected error",
		})
		return
	}
	c.JSON(http.StatusOK, cages)
}

func (api *API) GetCage(c *gin.Context) {
	cageLabel := c.Param("cageLabel")
	cage, err := api.parkManager.GetCage(cageLabel)
	if err != nil {
		if errors.Is(err, models.EntityNotFound) {
			c.JSON(http.StatusNotFound, models.ErrorResponse{
				ErrorMessage: fmt.Sprintf("cage with label %s not found", cageLabel),
			})
		} else {
			c.JSON(http.StatusInternalServerError, models.ErrorResponse{
				ErrorMessage: "unexpected error",
			})
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
		c.JSON(http.StatusUnprocessableEntity, models.ErrorResponse{
			ErrorMessage: "Request body is in the incorrect format",
		})
		return
	}

	err = api.parkManager.UpdateCagePowerStatus(cageLabel, updatePowerStatusRequest.HasPower)
	if err != nil {
		if errors.Is(err, models.IncompatibleCagePowerState) {
			c.JSON(http.StatusConflict, models.ErrorResponse{
				ErrorMessage: "the cage has dinosaurs in it and cannot be powered off",
			})
		} else if errors.Is(err, models.EntityNotFound) {
			c.JSON(http.StatusNotFound, models.ErrorResponse{
				ErrorMessage: "could not find cage",
			})
		} else {
			c.JSON(http.StatusInternalServerError, models.ErrorResponse{
				ErrorMessage: "unexpected error",
			})
		}
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "power status updated",
	})
}

func (api *API) GetDinosaursInCage(c *gin.Context) {
	cageLabel := c.Param("cageLabel")
	dinosaurs, err := api.parkManager.GetDinosaursInCage(cageLabel)
	if err != nil {
		if errors.Is(err, models.EntityNotFound) {
			c.JSON(http.StatusNotFound, models.ErrorResponse{
				ErrorMessage: fmt.Sprintf("the cage %s was not found", cageLabel),
			})
		} else {
			c.JSON(http.StatusInternalServerError, models.ErrorResponse{
				ErrorMessage: "unexpected error",
			})
		}
		return
	}
	c.JSON(http.StatusOK, dinosaurs)
}

func (api *API) AddDinosaurToCage(c *gin.Context) {
	var addDinosaurRequest models.AddDinosaurToCageRequest
	err := json.NewDecoder(c.Request.Body).Decode(&addDinosaurRequest)
	if err != nil {
		c.JSON(http.StatusUnprocessableEntity, models.ErrorResponse{
			ErrorMessage: "Request body is in the incorrect format",
		})
		return
	}
	targetCage := c.Param("cageLabel")
	err = api.parkManager.AddDinosaurToCage(addDinosaurRequest.Name, targetCage)
	if err != nil {
		if errors.Is(err, models.CageCapacityExceeded) {
			c.JSON(http.StatusConflict, models.ErrorResponse{
				ErrorMessage: "the cage is at capacity",
			})
		} else if errors.Is(err, models.IncompatibleCagePowerState) {
			c.JSON(http.StatusConflict, models.ErrorResponse{
				ErrorMessage: "the cage is unavailable at this time, because it does not have power",
			})
		} else if errors.Is(err, models.IncompatibleSpecies) {
			c.JSON(http.StatusConflict, models.ErrorResponse{
				ErrorMessage: "the cage already contains species of dinosaur that are incompatible with this dinosaur's specie",
			})
		} else if errors.Is(err, models.EntityNotFound) {
			c.JSON(http.StatusNotFound, models.ErrorResponse{
				ErrorMessage: "could not find either the cage or dinosaur",
			})
		} else {
			c.JSON(http.StatusInternalServerError, models.ErrorResponse{
				ErrorMessage: "unexpected error",
			})
		}
		return
	}
	c.JSON(http.StatusCreated, gin.H{
		"message": "dinosaur added",
	})
}

func (api *API) AddDinosaur(c *gin.Context) {
	var dinosaur models.Dinosaur
	err := json.NewDecoder(c.Request.Body).Decode(&dinosaur)
	if err != nil {
		c.JSON(http.StatusUnprocessableEntity, models.ErrorResponse{
			ErrorMessage: "Request body is in the incorrect format",
		})
		return
	}

	err = api.parkManager.AddDinosaur(dinosaur)
	if err != nil {
		if errors.Is(err, models.InvalidDinosaurSpecies) {
			c.JSON(http.StatusConflict, models.ErrorResponse{
				ErrorMessage: fmt.Sprintf("The species %s is not a valid dinosaur species", dinosaur.Species)})
		} else if errors.Is(err, models.EntityAlreadyExists) {
			c.JSON(http.StatusConflict, models.ErrorResponse{
				ErrorMessage: fmt.Sprintf("There is already a dinosaur with the name %s", dinosaur.Name),
			})
		}
		return
	}
	c.JSON(http.StatusCreated, dinosaur)
}

func (api *API) GetDinosaurs(c *gin.Context) {
	filter := models.DinosaurFilter{}
	if c.Query("species") != "" {
		species := c.Query("species")
		filter.Species = &species
	}
	if c.Query("diet") != "" {
		diet := c.Query("diet")
		filter.Diet = &diet
	}
	if c.Query("needsCageAssignment") != "" {
		needsCageAssignment := c.Query("needsCageAssignment") == "true"
		filter.NeedsCageAssignment = &needsCageAssignment
	}

	dinosaurs, err := api.parkManager.GetDinosaurs(filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			ErrorMessage: "unexpected error",
		})
		return
	}
	c.JSON(http.StatusOK, dinosaurs)
}

func (api *API) GetDinosaur(c *gin.Context) {
	dinosaurName := c.Param("name")
	dinosaur, err := api.parkManager.GetDinosaur(dinosaurName)
	if err != nil {
		if errors.Is(err, models.EntityNotFound) {
			c.JSON(http.StatusNotFound, models.ErrorResponse{
				ErrorMessage: fmt.Sprintf("dinosaur with name %s not found", dinosaurName),
			})
		} else {
			c.JSON(http.StatusInternalServerError, models.ErrorResponse{
				ErrorMessage: "unexpected error",
			})
		}
		return
	}
	c.JSON(http.StatusOK, dinosaur)
}
