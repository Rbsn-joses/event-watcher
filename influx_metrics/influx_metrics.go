package influx_metrics

import (
	"time"

	"github.com/influxdata/influxdb/client/v2"
)

func CreateMetrics(name string, namespace string, eventType string, objectType string, numericEvent float64) error {
	// Create a new HTTPClient
	c, err := client.NewHTTPClient(client.HTTPConfig{
		Addr:     "http://localhost:8086",
		Username: "teste",
		Password: "teste",
	})
	if err != nil {
		return err
	}
	defer c.Close()

	// Create a new point batch
	bp, err := client.NewBatchPoints(client.BatchPointsConfig{
		Database:  "teste",
		Precision: "s",
	})
	if err != nil {
		return err
	}
	// Create a point and add to batch
	tags := map[string]string{"object_name": name, "object_namespace": namespace, "object_eventType": eventType, "object_type": objectType}
	fields := map[string]interface{}{"event_number_identifier": numericEvent, "object_eventType": eventType, "object_type": objectType}

	pt, err := client.NewPoint("kubernetes_events", tags, fields, time.Now())
	if err != nil {
		return err
	}
	bp.AddPoint(pt)

	// Write the batch
	if err := c.Write(bp); err != nil {
		return err
	}

	// Close client resources
	if err := c.Close(); err != nil {
		return err
	}
	return err

}
