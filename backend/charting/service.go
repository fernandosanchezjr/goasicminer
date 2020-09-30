package charting

import (
	"errors"
	"github.com/fernandosanchezjr/goasicminer/backend/services/data"
	"github.com/fernandosanchezjr/goasicminer/backend/services/implementation"
	"github.com/fernandosanchezjr/goasicminer/networking/certs"
	"github.com/fernandosanchezjr/goasicminer/utils"
	"github.com/go-echarts/go-echarts/charts"
	log "github.com/sirupsen/logrus"
	"go.etcd.io/bbolt"
	"net/http"
	"strings"
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
	return http.ListenAndServeTLS(":8080", cert, key, cs)
}

func (cs *Service) GetHostName(request *http.Request) (string, error) {
	var parts []string
	for _, part := range strings.Split(request.URL.Path, "/") {
		if part != "" {
			parts = append(parts, part)
		}
	}
	if len(parts) == 1 {
		if found, err := implementation.IsHostNameKnown(cs.db, parts[0]); err != nil {
			return "", err
		} else if !found {
			return "", errors.New("invalid path")
		}
		return parts[0], nil
	} else {
		return "", errors.New("invalid path")
	}
}

func (cs *Service) GetMinutes() ([]*data.Minute, error) {
	end := time.Now()
	start := end.AddDate(0, 0, -1)
	return implementation.GetAggregatedMinutes(cs.db, "moria01", start, end)
}

func (cs *Service) GetChartData(minutes []*data.Minute) map[string]*Data {
	line := NewData()
	var serialNumber string
	for _, m := range minutes {
		serials := m.Events["serial"]
		difficulties := m.Events["difficulty"]
		for pos, t := range m.EventTimes {
			serialNumber = serials[pos].(string)
			line.Append(t, difficulties[pos], serialNumber)
		}
	}
	return line.SplitByLabels(utils.Difficulty(0))
}

func (cs *Service) BuildChart(lineData map[string]*Data) *charts.Line {
	lineChart := charts.NewLine()
	lineChart.SetGlobalOptions(
		charts.TitleOpts{Title: "Difficulty"},
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
	lineChart.AddJSFuncs("setTimeout(function(){location.reload();}, 60000);")
	return lineChart
}

func (cs *Service) ServeHTTP(w http.ResponseWriter, request *http.Request) {
	startTime := time.Now()
	hostName, err := cs.GetHostName(request)
	if err != nil {
		w.WriteHeader(404)
		return
	}
	minutes, err := cs.GetMinutes()
	if err != nil {
		w.WriteHeader(404)
		return
	}
	chartData := cs.GetChartData(minutes)
	lineChart := cs.BuildChart(chartData)
	if err := lineChart.Render(w); err != nil {
		log.WithError(err).Error("Error rendering chart")
	}
	log.WithFields(log.Fields{
		"elapsedTime":  time.Since(startTime),
		"totalMinutes": len(minutes),
		"path":         request.URL,
		"hostName":     hostName,
	}).Println("HTTP request")
}
