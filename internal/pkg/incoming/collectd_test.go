package incoming

import (
	"reflect"
	"strconv"
	"testing"
)

func GetFieldStr(dataItem DataTypeInterface, field string) string {
	r := reflect.ValueOf(dataItem)
	f := reflect.Indirect(r).FieldByName(field)
	return string(f.String())
}

func TestCollected(t *testing.T) {
	c := NewInComing(COLLECTD)
	jsonString := c.GenerateSampleJSON("hostname", "pluginname")
	if len(jsonString) == 0 {
		t.Error("Empty sample string generated")
	}

	c1 := NewInComing(COLLECTD)
	if len(GetFieldStr(c1, "Plugin")) != 0 {
		t.Error("Collectd data  is not empty.")
	}
	//2
	sample := c1.GenerateSampleData("hostname", "plugi_name")
	if len(GetFieldStr(sample, "Plugin")) == 0 {
		t.Errorf("Collectd data was not populated by GenrateSampleData %#v", c1)
	}
	c1 = NewInComing(COLLECTD)
	c1.ParseInputJSON(jsonString)
	if len(GetFieldStr(c1, "Plugin")) == 0 {
		t.Errorf("Collectd data was not populated by ParsestrconvInputJSON %#v", c1)
	}
	//check DSName method

	if collectd, ok := c1.(*Collectd); ok {
		for index := range collectd.Values {
			dsname := collectd.DSName(index)
			if len(dsname) == 0 {
				t.Errorf("Collectd DSName is empty %#v", dsname)
			}
		}
		//pass all DSname
		collectd.Dsnames = nil
		dsname := collectd.DSName(0)
		if dsname != strconv.FormatInt(int64(0), 10) {
			t.Errorf("Collectd DSName is not eq to value %s", strconv.FormatInt(int64(0), 10))
		}
		collectd.Values = []float64{1}
		dsname = collectd.DSName(0)
		if dsname != "value" {
			t.Errorf("Collectd DSName is not eq to value %s", dsname)
		}

	}

	c1 = NewInComing(COLLECTD)
	c1.ParseInputByte([]byte(jsonString))
	if len(GetFieldStr(c1, "Plugin")) == 0 {
		t.Errorf("Collectd data was not populated by ParseCollectdByte %#v", c1)
	}
	errors := c1.ParseInputByte([]byte("error string"))
	if errors == nil {
		t.Errorf("Excepted error got nil%#v", errors)
	}
}

func TestCollectedMetrics(t *testing.T) {
	c1 := NewInComing(COLLECTD)
	c := NewInComing(COLLECTD)
	jsonString := c.GenerateSampleJSON("hostname", "plugi_name")
	if len(jsonString) == 0 {
		t.Error("Empty sample string generated")
	}
	c1.ParseInputJSON(jsonString)
	if len(GetFieldStr(c1, "Plugin")) == 0 {
		t.Errorf("Collectd data was not populated by ParseInputJSON %#v", c1)
	}
	errors := c1.ParseInputJSON("Error Json")
	if errors == nil {
		t.Errorf("Excepted error got nil%v", errors)
	}
	labels := c1.GetLabels()
	if len(labels) < 2 {
		t.Errorf("Labels not populated by GetLabels %#v", c1)
	}
	name := c1.GetName()
	if len(name) == 0 {
		t.Errorf("name not populated by GetName %#v", c1)
	}
	metricDesc := c1.GetMetricDesc(0)
	if len(metricDesc) == 0 {
		t.Errorf("metricDesc not populated by GetMetricDesc %#v", c1)
	}
	metricName := c1.GetMetricName(0)
	if len(metricName) == 0 {
		t.Errorf("metricName not populated by GetMetricName %#v", c1)
	}

}
