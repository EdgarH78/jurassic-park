package integration_test

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/EdgarH78/jurassic-park/api"
	"github.com/EdgarH78/jurassic-park/data"
	"github.com/EdgarH78/jurassic-park/models"
	"github.com/gin-gonic/gin"
)

// We purposefully configure to a local database on a different port, because we don't interfere with a real database.
var config = data.SQLConfig{
	Host:         "localhost:3307",
	User:         "admin",
	Password:     "password",
	DatabaseName: "jurassicpark",
}

func TestCreateCage(t *testing.T) {
	cases := []struct {
		description        string
		cage               models.Cage
		expectedStatusCode int
	}{
		{
			description: "create cage",
			cage: models.Cage{
				Label:        "test-cage",
				MaxOccupancy: 10,
				HasPower:     true,
			},
			expectedStatusCode: http.StatusCreated,
		},
	}

	for _, c := range cases {
		t.Run(c.description, func(t *testing.T) {
			err := clearOutTestDatabase()
			if err != nil {
				t.Errorf("error when clearing out test database: %s", err)
				return
			}
			r := gin.Default()
			_, err = createTestApi(r)
			if err != nil {
				t.Errorf("error when creating test api: %s", err)
				return
			}

			cageBody, err := json.Marshal(c.cage)
			if err != nil {
				t.Errorf("unexpected error marshaling cage to json: %s", err)
				return
			}

			w := httptest.NewRecorder()

			req, _ := http.NewRequest("POST", "/jurassicpark/v1/cages", bytes.NewReader(cageBody))
			r.ServeHTTP(w, req)

			if w.Code != c.expectedStatusCode {
				t.Errorf("expected status code %d got %d", c.expectedStatusCode, w.Code)
			}
		})
	}
}

func TestGetCage(t *testing.T) {
	err := clearOutTestDatabase()
	if err != nil {
		t.Errorf("error when clearing out test database: %s", err)
		return
	}

	dao, err := data.NewParkSqlDao(config)
	if err != nil {
		t.Errorf("error when creating test dao: %s", err)
	}
	err = dao.AddCage(models.Cage{
		Label:        "test-cage-1",
		MaxOccupancy: 10,
		HasPower:     true,
	})

	if err != nil {
		t.Errorf("error when creating test cage: %s", err)
	}

	cases := []struct {
		description        string
		cageLabel          string
		expectedCage       *models.Cage
		expectedStatusCode int
	}{
		{
			description: "get cage, cage is returned",
			cageLabel:   "test-cage-1",
			expectedCage: &models.Cage{
				Label:        "test-cage-1",
				MaxOccupancy: 10,
				HasPower:     true,
				Occupancy:    0,
			},
			expectedStatusCode: http.StatusOK,
		},
		{
			description:        "get cage, cage is returned",
			cageLabel:          "abc",
			expectedCage:       nil,
			expectedStatusCode: http.StatusNotFound,
		},
	}

	for _, c := range cases {
		t.Run(c.description, func(t *testing.T) {
			r := gin.Default()
			_, err = createTestApi(r)
			if err != nil {
				t.Errorf("error when creating test api: %s", err)
				return
			}

			w := httptest.NewRecorder()

			req, _ := http.NewRequest("GET", fmt.Sprintf("/jurassicpark/v1/cages/%s", c.cageLabel), nil)

			r.ServeHTTP(w, req)

			if w.Code != c.expectedStatusCode {
				t.Errorf("expected status code %d got %d", c.expectedStatusCode, w.Code)
			}
			if c.expectedCage != nil {
				var actualCage models.Cage
				err = json.NewDecoder(w.Result().Body).Decode(&actualCage)
				if err != nil {
					t.Errorf("error while decoding result body: %s", err)
					return
				}
				assertCagesMatch(*c.expectedCage, actualCage, t)
			}
		})
	}
}

func TestGetCages(t *testing.T) {
	err := clearOutTestDatabase()
	if err != nil {
		t.Errorf("error when clearing out test database: %s", err)
		return
	}

	dao, err := data.NewParkSqlDao(config)
	if err != nil {
		t.Errorf("error when creating test dao: %s", err)
	}
	err = dao.AddCage(models.Cage{
		Label:        "test-cage-1",
		MaxOccupancy: 10,
		HasPower:     true,
	})
	if err != nil {
		t.Errorf("error when creating test cage: %s", err)
	}
	err = dao.AddCage(models.Cage{
		Label:        "test-cage-2",
		MaxOccupancy: 7,
		HasPower:     false,
	})

	if err != nil {
		t.Errorf("error when creating test cage: %s", err)
	}

	cases := []struct {
		description        string
		expectedCages      []models.Cage
		expectedStatusCode int
	}{
		{
			description: "cages are returned",
			expectedCages: []models.Cage{
				{
					Label:        "test-cage-1",
					MaxOccupancy: 10,
					HasPower:     true,
					Occupancy:    0,
				},
				{
					Label:        "test-cage-2",
					MaxOccupancy: 7,
					HasPower:     false,
					Occupancy:    0,
				},
			},
			expectedStatusCode: http.StatusOK,
		},
	}
	for _, c := range cases {
		t.Run(c.description, func(t *testing.T) {
			r := gin.Default()
			_, err = createTestApi(r)
			if err != nil {
				t.Errorf("error when creating test api: %s", err)
				return
			}

			w := httptest.NewRecorder()

			req, _ := http.NewRequest("GET", "/jurassicpark/v1/cages", nil)

			r.ServeHTTP(w, req)

			if w.Code != c.expectedStatusCode {
				t.Errorf("expected status code %d got %d", c.expectedStatusCode, w.Code)
			}
			if c.expectedCages != nil && len(c.expectedCages) > 0 {
				var actualCages []models.Cage
				err = json.NewDecoder(w.Result().Body).Decode(&actualCages)
				if err != nil {
					t.Errorf("error while decoding result body: %s", err)
					return
				}
				if len(actualCages) != len(c.expectedCages) {
					t.Errorf("expected %d cages to be returned got %d", len(c.expectedCages), len(actualCages))
				}

				for i := 0; i < len(c.expectedCages); i++ {
					expectedCage := c.expectedCages[i]
					actualCage := actualCages[i]
					assertCagesMatch(expectedCage, actualCage, t)
				}
			}
		})
	}
}

func assertCagesMatch(expectedCage, actualCage models.Cage, t *testing.T) {
	if expectedCage.Label != actualCage.Label {
		t.Errorf("expected cage label to be %s got %s", expectedCage.Label, actualCage.Label)
	}
	if expectedCage.HasPower != actualCage.HasPower {
		t.Errorf("expected cage HasPower to be %t got %t", expectedCage.HasPower, actualCage.HasPower)
	}
	if expectedCage.MaxOccupancy != actualCage.MaxOccupancy {
		t.Errorf("expected cage MaxOccupancy to be %d got %d", expectedCage.MaxOccupancy, actualCage.MaxOccupancy)
	}
	if expectedCage.Occupancy != actualCage.Occupancy {
		t.Errorf("expected cage Occupancy to be %d got %d", expectedCage.Occupancy, actualCage.Occupancy)
	}
}

func TestAddDinosaur(t *testing.T) {

	err := clearOutTestDatabase()
	if err != nil {
		t.Errorf("error when clearing out test database: %s", err)
		return
	}

	cases := []struct {
		description        string
		dinosaur           models.Dinosaur
		expectedStatusCode int
	}{
		{
			description: "add Velociraptor",
			dinosaur: models.Dinosaur{
				Name:    "Vela",
				Species: "Velociraptor",
			},
			expectedStatusCode: http.StatusCreated,
		},
		{
			description: "add Brachiosaurus",
			dinosaur: models.Dinosaur{
				Name:    "Brachen",
				Species: "Brachiosaurus",
			},
			expectedStatusCode: http.StatusCreated,
		},
		{
			description: "add Brachiosaurus",
			dinosaur: models.Dinosaur{
				Name:    "Brachen",
				Species: "Triceratops",
			},
			expectedStatusCode: http.StatusConflict,
		},
		{
			description: "add Mythosaurus",
			dinosaur: models.Dinosaur{
				Name:    "Mythos",
				Species: "Mythosaurus",
			},
			expectedStatusCode: http.StatusConflict,
		},
	}

	for _, c := range cases {
		t.Run(c.description, func(t *testing.T) {
			r := gin.Default()
			_, err := createTestApi(r)
			if err != nil {
				t.Errorf("error when creating test api: %s", err)
				return
			}

			dinoBody, err := json.Marshal(c.dinosaur)
			if err != nil {
				t.Errorf("unexpected error marshaling cage to json: %s", err)
				return
			}

			w := httptest.NewRecorder()

			req, _ := http.NewRequest("POST", "/jurassicpark/v1/dinosaurs", bytes.NewReader(dinoBody))
			r.ServeHTTP(w, req)

			if w.Code != c.expectedStatusCode {
				t.Errorf("expected status code %d got %d", c.expectedStatusCode, w.Code)
			}
		})
	}
}

func TestGetDinosaurs(t *testing.T) {
	err := clearOutTestDatabase()
	if err != nil {
		t.Errorf("error when clearing out test database: %s", err)
		return
	}

	dao, err := data.NewParkSqlDao(config)
	if err != nil {
		t.Errorf("error when creating test dao: %s", err)
		return
	}

	err = dao.AddDinosaur(models.Dinosaur{
		Name:    "Vela",
		Species: "Velociraptor",
	})
	if err != nil {
		t.Errorf("error when adding dinosaur")
		return
	}
	err = dao.AddDinosaur(models.Dinosaur{
		Name:    "Brachen",
		Species: "Brachiosaurus",
	})
	if err != nil {
		t.Errorf("error when adding dinosaur")
		return
	}

	cases := []struct {
		description        string
		expectedStatusCode int
		expectedDinosaurs  []models.Dinosaur
	}{
		{
			description:        "get dinosaurs returns dinosaurs",
			expectedStatusCode: http.StatusOK,
			expectedDinosaurs: []models.Dinosaur{
				{
					Name:    "Vela",
					Species: "Velociraptor",
					Diet:    "Carnivore",
					Cage:    nil,
				},
				{
					Name:    "Brachen",
					Species: "Brachiosaurus",
					Diet:    "Herbivore",
					Cage:    nil,
				},
			},
		},
	}

	for _, c := range cases {
		t.Run(c.description, func(t *testing.T) {
			r := gin.Default()
			_, err = createTestApi(r)
			if err != nil {
				t.Errorf("error when creating test api: %s", err)
				return
			}

			w := httptest.NewRecorder()

			req, _ := http.NewRequest("GET", "/jurassicpark/v1/dinosaurs", nil)

			r.ServeHTTP(w, req)

			if w.Code != c.expectedStatusCode {
				t.Errorf("expected status code %d got %d", c.expectedStatusCode, w.Code)
			}
			if c.expectedDinosaurs != nil && len(c.expectedDinosaurs) > 0 {
				var actualDinosaurs []models.Dinosaur
				err = json.NewDecoder(w.Result().Body).Decode(&actualDinosaurs)
				if err != nil {
					t.Errorf("error while decoding result body: %s", err)
					return
				}
				if len(actualDinosaurs) != len(c.expectedDinosaurs) {
					t.Errorf("expected %d cages to be returned got %d", len(c.expectedDinosaurs), len(actualDinosaurs))
				}

				for i := 0; i < len(c.expectedDinosaurs); i++ {
					expectedDinosaur := c.expectedDinosaurs[i]
					actualDinosaur := actualDinosaurs[i]
					assertDinosaursMatch(expectedDinosaur, actualDinosaur, t)
				}
			}
		})
	}
}

func TestGetDinosaur(t *testing.T) {
	err := clearOutTestDatabase()
	if err != nil {
		t.Errorf("error when clearing out test database: %s", err)
		return
	}

	dao, err := data.NewParkSqlDao(config)
	if err != nil {
		t.Errorf("error when creating test dao: %s", err)
		return
	}

	err = dao.AddDinosaur(models.Dinosaur{
		Name:    "Vela",
		Species: "Velociraptor",
	})

	cases := []struct {
		description        string
		dinosaurName       string
		expectedDinosaur   *models.Dinosaur
		expectedStatusCode int
	}{
		{
			description:        "get Vela the Velociraptor",
			dinosaurName:       "Vela",
			expectedStatusCode: http.StatusOK,
			expectedDinosaur: &models.Dinosaur{
				Name:    "Vela",
				Species: "Velociraptor",
				Diet:    "Carnivore",
				Cage:    nil,
			},
		},
		{
			description:        "get a dinosaur that is not at the park",
			dinosaurName:       "NotHere",
			expectedStatusCode: http.StatusNotFound,
		},
	}

	for _, c := range cases {
		t.Run(c.description, func(t *testing.T) {
			r := gin.Default()
			_, err = createTestApi(r)
			if err != nil {
				t.Errorf("error when creating test api: %s", err)
				return
			}

			w := httptest.NewRecorder()

			req, _ := http.NewRequest("GET", fmt.Sprintf("/jurassicpark/v1/dinosaurs/%s", c.dinosaurName), nil)

			r.ServeHTTP(w, req)

			if w.Code != c.expectedStatusCode {
				t.Errorf("expected status code %d got %d", c.expectedStatusCode, w.Code)
			}
			if c.expectedDinosaur != nil {
				var actualDinosaur models.Dinosaur
				err = json.NewDecoder(w.Result().Body).Decode(&actualDinosaur)
				if err != nil {
					t.Errorf("error while decoding result body: %s", err)
					return
				}
				assertDinosaursMatch(*c.expectedDinosaur, actualDinosaur, t)
			}
		})
	}
}

func TestAddDinosaurToCage(t *testing.T) {
	err := clearOutTestDatabase()
	if err != nil {
		t.Errorf("error when clearing out test database: %s", err)
		return
	}

	dao, err := data.NewParkSqlDao(config)
	if err != nil {
		t.Errorf("error when creating test dao: %s", err)
		return
	}

	dao.AddCage(models.Cage{
		Label:        "T-Rex-Pen",
		MaxOccupancy: 2,
		HasPower:     true,
	})
	dao.AddCage(models.Cage{
		Label:        "Raptor-Pen-1",
		MaxOccupancy: 10,
		HasPower:     true,
	})
	dao.AddCage(models.Cage{
		Label:        "Raptor-Pen-2",
		MaxOccupancy: 5,
		HasPower:     false,
	})
	dao.AddCage(models.Cage{
		Label:        "Herbivore-Pen",
		MaxOccupancy: 10,
		HasPower:     true,
	})

	dao.AddDinosaur(models.Dinosaur{
		Name:    "TerryRex",
		Species: "Tyrannosaurus",
	})
	dao.AddDinosaur(models.Dinosaur{
		Name:    "MerryRex",
		Species: "Tyrannosaurus",
	})
	dao.AddDinosaur(models.Dinosaur{
		Name:    "JerryRex",
		Species: "Tyrannosaurus",
	})
	dao.AddDinosaur(models.Dinosaur{
		Name:    "Vela",
		Species: "Velociraptor",
	})
	dao.AddDinosaur(models.Dinosaur{
		Name:    "Verona",
		Species: "Velociraptor",
	})
	dao.AddDinosaur(models.Dinosaur{
		Name:    "Talon",
		Species: "Verlociraptor",
	})
	dao.AddDinosaur(models.Dinosaur{
		Name:    "LittleFoot",
		Species: "Brachiosaurus",
	})
	dao.AddDinosaur(models.Dinosaur{
		Name:    "Cera",
		Species: "Triceratops",
	})
	dao.AddDinosaur(models.Dinosaur{
		Name:    "Rooter",
		Species: "Stegosaurus",
	})

	cases := []struct {
		description        string
		dinosaurToAdd      string
		targetCage         string
		expectedStatusCode int
	}{
		{
			description:        "add terry to T-Rex-Pen",
			dinosaurToAdd:      "TerryRex",
			targetCage:         "T-Rex-Pen",
			expectedStatusCode: http.StatusCreated,
		},
		{
			description:        "add Vela to T-Rex-Pen",
			dinosaurToAdd:      "Vela",
			targetCage:         "T-Rex-Pen",
			expectedStatusCode: http.StatusConflict,
		},
		{
			description:        "add MerryRex to T-Rex-Pen",
			dinosaurToAdd:      "MerryRex",
			targetCage:         "T-Rex-Pen",
			expectedStatusCode: http.StatusCreated,
		},
		{
			description:        "add JerryRex to T-Rex-Pen",
			dinosaurToAdd:      "JerryRex",
			targetCage:         "T-Rex-Pen",
			expectedStatusCode: http.StatusConflict,
		},
		{
			description:        "add Vela to Raptor-Pen-2",
			dinosaurToAdd:      "Vela",
			targetCage:         "Raptor-Pen-2",
			expectedStatusCode: http.StatusConflict,
		},
		{
			description:        "add Vela to Raptor-Pen-1",
			dinosaurToAdd:      "Vela",
			targetCage:         "Raptor-Pen-1",
			expectedStatusCode: http.StatusCreated,
		},
		{
			description:        "add LittleFoot to Raptor-Pen-1",
			dinosaurToAdd:      "LittleFoot",
			targetCage:         "Raptor-Pen-1",
			expectedStatusCode: http.StatusConflict,
		},
		{
			description:        "add LittleFoot to Herbivore-Pen",
			dinosaurToAdd:      "LittleFoot",
			targetCage:         "Herbivore-Pen",
			expectedStatusCode: http.StatusCreated,
		},
		{
			description:        "add Cera to Herbivore-Pen",
			dinosaurToAdd:      "Cera",
			targetCage:         "Herbivore-Pen",
			expectedStatusCode: http.StatusCreated,
		},
		{
			description:        "add Unkown dinosaur to Herbivore-Pen",
			dinosaurToAdd:      "HasNotBeenAdded",
			targetCage:         "Herbivore-Pen",
			expectedStatusCode: http.StatusNotFound,
		},
		{
			description:        "add Rooter to missing Pen",
			dinosaurToAdd:      "HasNotBeenAdded",
			targetCage:         "MisingPen",
			expectedStatusCode: http.StatusNotFound,
		},
	}

	for _, c := range cases {
		t.Run(c.description, func(t *testing.T) {
			r := gin.Default()
			_, err = createTestApi(r)
			if err != nil {
				t.Errorf("error when creating test api: %s", err)
				return
			}

			w := httptest.NewRecorder()

			addRequest := models.AddDinosaurToCageRequest{
				Name: c.dinosaurToAdd,
			}

			addBody, err := json.Marshal(addRequest)
			if err != nil {
				t.Errorf("failed to marshal add request: %s", err)
				return
			}
			req, _ := http.NewRequest("POST", fmt.Sprintf("/jurassicpark/v1/cages/%s/dinosaurs", c.targetCage), bytes.NewReader(addBody))

			r.ServeHTTP(w, req)

			if w.Code != c.expectedStatusCode {
				t.Errorf("expected status code %d got %d", c.expectedStatusCode, w.Code)
			}
		})
	}
}

func TestGetDinosaursForCage(t *testing.T) {
	err := clearOutTestDatabase()
	if err != nil {
		t.Errorf("error when clearing out test database: %s", err)
		return
	}

	dao, err := data.NewParkSqlDao(config)
	if err != nil {
		t.Errorf("error when creating test dao: %s", err)
		return
	}

	dao.AddCage(models.Cage{
		Label:        "T-Rex-Pen",
		MaxOccupancy: 2,
		HasPower:     true,
	})
	dao.AddDinosaur(models.Dinosaur{
		Name:    "TerryRex",
		Species: "Tyrannosaurus",
	})
	dao.AddDinosaur(models.Dinosaur{
		Name:    "MerryRex",
		Species: "Tyrannosaurus",
	})
	dao.AddDinosaur(models.Dinosaur{
		Name:    "LittleFoot",
		Species: "Brachiosaurus",
	})
	dao.AddDinosaurToCage("TerryRex", "T-Rex-Pen")
	dao.AddDinosaurToCage("MerryRex", "T-Rex-Pen")

	cases := []struct {
		description        string
		cageLabel          string
		expectedDinosaurs  []models.Dinosaur
		expectedStatusCode int
	}{
		{
			description:        "get dinosaurs in the T-Rex-Pen",
			cageLabel:          "T-Rex-Pen",
			expectedStatusCode: http.StatusOK,
			expectedDinosaurs: []models.Dinosaur{
				{
					Name:    "TerryRex",
					Species: "Tyrannosaurus",
					Diet:    "Carnivore",
					Cage:    wrapString("T-Rex-Pen"),
				},
				{
					Name:    "MerryRex",
					Species: "Tyrannosaurus",
					Diet:    "Carnivore",
					Cage:    wrapString("T-Rex-Pen"),
				},
			},
		},
		{
			description:        "get dinosaurs in missing cage",
			cageLabel:          "BadLabel",
			expectedStatusCode: http.StatusNotFound,
		},
	}

	for _, c := range cases {
		t.Run(c.description, func(t *testing.T) {
			r := gin.Default()
			_, err = createTestApi(r)
			if err != nil {
				t.Errorf("error when creating test api: %s", err)
				return
			}

			w := httptest.NewRecorder()

			req, _ := http.NewRequest("GET", fmt.Sprintf("/jurassicpark/v1/cages/%s/dinosaurs", c.cageLabel), nil)

			r.ServeHTTP(w, req)

			if w.Code != c.expectedStatusCode {
				t.Errorf("expected status code %d got %d", c.expectedStatusCode, w.Code)
			}
			if c.expectedDinosaurs != nil && len(c.expectedDinosaurs) > 0 {
				var actualDinosaurs []models.Dinosaur
				err = json.NewDecoder(w.Result().Body).Decode(&actualDinosaurs)
				if err != nil {
					t.Errorf("error while decoding result body: %s", err)
					return
				}
				if len(actualDinosaurs) != len(c.expectedDinosaurs) {
					t.Errorf("expected %d cages to be returned got %d", len(c.expectedDinosaurs), len(actualDinosaurs))
				}

				for i := 0; i < len(c.expectedDinosaurs); i++ {
					expectedDinosaur := c.expectedDinosaurs[i]
					actualDinosaur := actualDinosaurs[i]
					assertDinosaursMatch(expectedDinosaur, actualDinosaur, t)
				}
			}
		})
	}
}

func TestSwitchPowerStatusOnCages(t *testing.T) {
	err := clearOutTestDatabase()
	if err != nil {
		t.Errorf("error when clearing out test database: %s", err)
		return
	}

	dao, err := data.NewParkSqlDao(config)
	if err != nil {
		t.Errorf("error when creating test dao: %s", err)
		return
	}

	dao.AddCage(models.Cage{
		Label:        "T-Rex-Pen",
		MaxOccupancy: 2,
		HasPower:     true,
	})
	dao.AddCage(models.Cage{
		Label:        "Raptor-Pen-1",
		MaxOccupancy: 5,
		HasPower:     true,
	})
	dao.AddDinosaur(models.Dinosaur{
		Name:    "TerryRex",
		Species: "Tyrannosaurus",
	})
	dao.AddDinosaur(models.Dinosaur{
		Name:    "MerryRex",
		Species: "Tyrannosaurus",
	})

	dao.AddDinosaurToCage("TerryRex", "T-Rex-Pen")
	dao.AddDinosaurToCage("MerryRex", "T-Rex-Pen")

	cases := []struct {
		description        string
		targetCage         string
		powerOn            bool
		expectedStatusCode int
	}{
		{
			description:        "turn off power to Raptor-Pen-1",
			targetCage:         "Raptor-Pen-1",
			powerOn:            false,
			expectedStatusCode: http.StatusOK,
		},
		{
			description:        "turn on power for Raptor-Pen-1",
			targetCage:         "Raptor-Pen-1",
			powerOn:            true,
			expectedStatusCode: http.StatusOK,
		},
		{
			description:        "turn off power to T-Rex-Pen",
			targetCage:         "T-Rex-Pen",
			powerOn:            false,
			expectedStatusCode: http.StatusConflict,
		},
		{
			description:        "turn off power a cage that does not exist",
			targetCage:         "NotHere",
			powerOn:            false,
			expectedStatusCode: http.StatusNotFound,
		},
	}

	for _, c := range cases {
		t.Run(c.description, func(t *testing.T) {
			r := gin.Default()
			_, err = createTestApi(r)
			if err != nil {
				t.Errorf("error when creating test api: %s", err)
				return
			}

			w := httptest.NewRecorder()

			powerChangeRequest := models.UpdateCagePowerStatusRequest{
				HasPower: c.powerOn,
			}

			powerChangeBody, err := json.Marshal(powerChangeRequest)
			if err != nil {
				t.Errorf("failed to marshal add request: %s", err)
				return
			}
			req, _ := http.NewRequest("PATCH", fmt.Sprintf("/jurassicpark/v1/cages/%s", c.targetCage), bytes.NewReader(powerChangeBody))

			r.ServeHTTP(w, req)

			if w.Code != c.expectedStatusCode {
				t.Errorf("expected status code %d got %d", c.expectedStatusCode, w.Code)
			}
		})
	}
}

func wrapString(value string) *string {
	return &value
}

func assertDinosaursMatch(expectedDinosaur, actualDinosaur models.Dinosaur, t *testing.T) {
	if expectedDinosaur.Name != actualDinosaur.Name {
		t.Errorf("expected dinosaur Name to be %s got %s", expectedDinosaur.Name, actualDinosaur.Name)
	}
	if expectedDinosaur.Species != actualDinosaur.Species {
		t.Errorf("expected dinosaur Species to be %s got %s", expectedDinosaur.Species, actualDinosaur.Species)
	}
	if expectedDinosaur.Diet != actualDinosaur.Diet {
		t.Errorf("expected dinosaur Diet to be %s got %s", expectedDinosaur.Diet, actualDinosaur.Diet)
	}
	if expectedDinosaur.Cage != nil || actualDinosaur.Cage != nil {
		if expectedDinosaur.Cage == nil && actualDinosaur.Cage != nil {
			t.Errorf("expected dinosaur Cage should be nil got %s", *actualDinosaur.Cage)
		} else if expectedDinosaur.Cage != nil && actualDinosaur.Cage == nil {
			t.Errorf("expected dinosaur Cage to be %s got nil", *expectedDinosaur.Cage)
		} else if *expectedDinosaur.Cage != *actualDinosaur.Cage {
			t.Errorf("expected dinosaur Cage to %s got %s", *expectedDinosaur.Cage, *actualDinosaur.Cage)
		}
	}
}

func createTestApi(engine *gin.Engine) (*api.API, error) {
	dao, err := data.NewParkSqlDao(config)
	if err != nil {
		return nil, err
	}

	api := api.NewAPI(dao, engine)
	return api, nil
}

func clearOutTestDatabase() error {
	db, err := sql.Open("mysql", config.ConnectionString())
	if err != nil {
		return err
	}

	_, err = db.Exec("TRUNCATE dinosaur")
	if err != nil {
		return err
	}

	_, err = db.Exec("DELETE FROM cage")
	if err != nil {
		return err
	}
	return nil
}
