package charting

import (
	"github.com/fernandosanchezjr/goasicminer/backend/services/data"
	"github.com/fernandosanchezjr/goasicminer/backend/services/implementation"
	"github.com/fernandosanchezjr/goasicminer/networking/certs"
	"github.com/fernandosanchezjr/goasicminer/utils"
	"github.com/go-echarts/go-echarts/charts"
	"github.com/julienschmidt/httprouter"
	log "github.com/sirupsen/logrus"
	"go.etcd.io/bbolt"
	"net/http"
	"time"
)

type Service struct {
	db *bbolt.DB
}

func NewService(db *bbolt.DB) *Service {
	return &Service{db: db}
}

func (cs *Service) Start() error {
	cert, key, _ := certs.GetCertsPath("https")
	router := httprouter.New()
	router.GET("/difficulty/:hostName", cs.GetHostNamePerformance)
	return http.ListenAndServeTLS(":8080", cert, key, router)
}

func (cs *Service) TimeSlices(days, hours, minutes int) ([]*data.Minute, error) {
	end := time.Now()
	start := end.Add(
		-time.Duration(days) * 24 * time.Hour,
	).Add(
		-time.Duration(hours) * time.Hour,
	).Add(
		-time.Duration(minutes) * time.Minute,
	)
	return implementation.GetAggregatedMinutes(cs.db, "moria01", start, end)
}

func (cs *Service) GetChartData(timeSlices []*data.Minute) map[string]*Data {
	line := NewData()
	var serialNumber string
	for _, m := range timeSlices {
		serials := m.Events["serial"]
		difficulties := m.Events["difficulty"]
		for pos, t := range m.EventTimes {
			serialNumber = serials[pos].(string)
			line.Append(t, difficulties[pos], serialNumber)
		}
	}
	return line.SplitByLabels(utils.Difficulty(0))
}

func (cs *Service) BuildChart(lineData map[string]*Data, refresh bool) *charts.Line {
	lineChart := charts.NewLine()
	lineChart.SetGlobalOptions(
		charts.InitOpts{
			Width:  "100wh",
			Height: "85vh",
		},
		charts.TitleOpts{Title: "Difficulty, by device"},
		charts.ToolboxOpts{
			Show: true,
		},
	)
	pos := 0
	for label, line := range lineData {
		if pos == 0 {
			lineChart.AddXAxis(line.X)
		}
		pos += 1
		lineChart.AddYAxis(label, line.Y)
	}
	if refresh {
		lineChart.AddJSFuncs("setTimeout(function(){location.reload();}, 60000);")
	}
	return lineChart
}

func (cs *Service) GetHostNamePerformance(
	w http.ResponseWriter,
	request *http.Request,
	params httprouter.Params,
) {
	var err error
	var serviceParams *ServiceParams
	startTime := time.Now()
	hostName := params.ByName("hostName")
	if found, err := implementation.IsHostNameKnown(cs.db, hostName); err != nil {
		log.WithError(err).Error("Error checking hostname")
		w.WriteHeader(404)
		return
	} else if !found {
		log.WithField("hostName", hostName).Error("Hostname not found")
		w.WriteHeader(404)
		return
	}
	query := request.URL.Query()
	if serviceParams, err = ParseServiceParams(query); err != nil {
		w.WriteHeader(404)
		return
	}
	timeSlices, err := cs.TimeSlices(serviceParams.Days, serviceParams.Hours, serviceParams.Minutes)
	if err != nil {
		w.WriteHeader(404)
		return
	}
	chartData := cs.GetChartData(timeSlices)
	lineChart := cs.BuildChart(chartData, serviceParams.Refresh)
	if err := lineChart.Render(w); err != nil {
		log.WithError(err).Error("Error rendering chart")
	}
	log.WithFields(log.Fields{
		"elapsedTime":  time.Since(startTime),
		"totalMinutes": len(timeSlices),
		"path":         request.URL,
		"hostName":     hostName,
	}).Println("Chart request")
}
