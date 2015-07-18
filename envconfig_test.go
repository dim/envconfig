package envconfig_test

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/vrischmann/envconfig"
)

func TestParseSimpleConfig(t *testing.T) {
	var conf struct {
		Name string
		Log  struct {
			Path string
		}
	}

	err := envconfig.Init(&conf)
	require.Equal(t, "envconfig: keys NAME, name not found", err.Error())

	os.Setenv("NAME", "foobar")
	err = envconfig.Init(&conf)
	require.Equal(t, "envconfig: keys LOG_PATH, log_path not found", err.Error())

	os.Setenv("LOG_PATH", "/var/log/foobar")
	err = envconfig.Init(&conf)
	require.Nil(t, err)

	require.Equal(t, "foobar", conf.Name)
	require.Equal(t, "/var/log/foobar", conf.Log.Path)

	// Clean up at the end of the test - some tests share the same key and we don't values to be seen by those tests
	os.Setenv("NAME", "")
	os.Setenv("LOG_PATH", "")
}

func TestParseIntegerConfig(t *testing.T) {
	var conf struct {
		Port    int
		Long    uint64
		Version uint8
	}

	timestamp := time.Now().UnixNano()

	os.Setenv("PORT", "80")
	os.Setenv("LONG", fmt.Sprintf("%d", timestamp))
	os.Setenv("VERSION", "2")

	err := envconfig.Init(&conf)
	require.Nil(t, err)

	require.Equal(t, 80, conf.Port)
	require.Equal(t, uint64(timestamp), conf.Long)
	require.Equal(t, uint8(2), conf.Version)
}

func TestParseBoolConfig(t *testing.T) {
	var conf struct {
		DoIt bool
	}

	os.Setenv("DOIT", "true")

	err := envconfig.Init(&conf)
	require.Nil(t, err)
	require.Equal(t, true, conf.DoIt)
}

func TestParseBytesConfig(t *testing.T) {
	var conf struct {
		Data []byte
	}

	os.Setenv("DATA", "Rk9PQkFS")

	err := envconfig.Init(&conf)
	require.Nil(t, err)
	require.Equal(t, []byte("FOOBAR"), conf.Data)
}

func TestParseFloatConfig(t *testing.T) {
	var conf struct {
		Delta  float32
		DeltaV float64
	}

	os.Setenv("DELTA", "0.02")
	os.Setenv("DELTAV", "400.20000000001")

	err := envconfig.Init(&conf)
	require.Nil(t, err)
	require.Equal(t, float32(0.02), conf.Delta)
	require.Equal(t, float64(400.20000000001), conf.DeltaV)
}

func TestParseSliceConfig(t *testing.T) {
	var conf struct {
		Names  []string
		Ports  []int
		Shards []struct {
			Name string
			Addr string
		}
	}

	os.Setenv("NAMES", "foobar,barbaz")
	os.Setenv("PORTS", "900,100")
	os.Setenv("SHARDS", "{foobar,localhost:2929},{barbaz,localhost:2828}")

	err := envconfig.Init(&conf)
	require.Nil(t, err)

	require.Equal(t, 2, len(conf.Names))
	require.Equal(t, "foobar", conf.Names[0])
	require.Equal(t, "barbaz", conf.Names[1])
	require.Equal(t, 2, len(conf.Ports))
	require.Equal(t, 900, conf.Ports[0])
	require.Equal(t, 100, conf.Ports[1])
	require.Equal(t, 2, len(conf.Shards))
	require.Equal(t, "foobar", conf.Shards[0].Name)
	require.Equal(t, "localhost:2929", conf.Shards[0].Addr)
	require.Equal(t, "barbaz", conf.Shards[1].Name)
	require.Equal(t, "localhost:2828", conf.Shards[1].Addr)
}

func TestDurationConfig(t *testing.T) {
	var conf struct {
		Timeout time.Duration
	}

	os.Setenv("TIMEOUT", "1m")

	err := envconfig.Init(&conf)
	require.Nil(t, err)

	require.Equal(t, time.Minute*1, conf.Timeout)
}

func TestInvalidDurationConfig(t *testing.T) {
	var conf struct {
		Timeout time.Duration
	}

	os.Setenv("TIMEOUT", "foo")

	err := envconfig.Init(&conf)
	require.NotNil(t, err)
}

func TestAllPointerConfig(t *testing.T) {
	var conf struct {
		Name   *string
		Port   *int
		Delta  *float32
		DeltaV *float64
		Hosts  *[]string
		Shards *[]*struct {
			Name *string
			Addr *string
		}
		Master *struct {
			Name *string
			Addr *string
		}
		Timeout *time.Duration
	}

	os.Setenv("NAME", "foobar")
	os.Setenv("PORT", "9000")
	os.Setenv("DELTA", "40.01")
	os.Setenv("DELTAV", "200.00001")
	os.Setenv("HOSTS", "localhost,free.fr")
	os.Setenv("SHARDS", "{foobar,localhost:2828},{barbaz,localhost:2929}")
	os.Setenv("MASTER_NAME", "master")
	os.Setenv("MASTER_ADDR", "localhost:2727")
	os.Setenv("TIMEOUT", "1m")

	err := envconfig.Init(&conf)
	require.Nil(t, err)

	require.Equal(t, "foobar", *conf.Name)
	require.Equal(t, 9000, *conf.Port)
	require.Equal(t, float32(40.01), *conf.Delta)
	require.Equal(t, 200.00001, *conf.DeltaV)
	require.Equal(t, 2, len(*conf.Hosts))
	require.Equal(t, "localhost", (*conf.Hosts)[0])
	require.Equal(t, "free.fr", (*conf.Hosts)[1])
	require.Equal(t, 2, len(*conf.Shards))
	require.Equal(t, "foobar", *(*conf.Shards)[0].Name)
	require.Equal(t, "localhost:2828", *(*conf.Shards)[0].Addr)
	require.Equal(t, "barbaz", *(*conf.Shards)[1].Name)
	require.Equal(t, "localhost:2929", *(*conf.Shards)[1].Addr)
	require.Equal(t, "master", *conf.Master.Name)
	require.Equal(t, "localhost:2727", *conf.Master.Addr)
	require.Equal(t, time.Minute*1, *conf.Timeout)
}

type logMode uint

const (
	logFile logMode = iota + 1
	logStdout
)

func (m *logMode) Unmarshal(s string) error {
	switch strings.ToLower(s) {
	case "file":
		*m = logFile
	case "stdout":
		*m = logStdout
	default:
		return fmt.Errorf("unable to unmarshal %s", s)
	}

	return nil
}

func TestUnmarshaler(t *testing.T) {
	var conf struct {
		LogMode logMode
	}

	os.Setenv("LOGMODE", "file")

	err := envconfig.Init(&conf)
	require.Nil(t, err)

	require.Equal(t, logFile, conf.LogMode)
}

func TestParseOptionalConfig(t *testing.T) {
	var conf struct {
		Name    string        `envconfig:"optional"`
		Flag    bool          `envconfig:"optional"`
		Timeout time.Duration `envconfig:"optional"`
		Port    int           `envconfig:"optional"`
		Port2   uint          `envconfig:"optional"`
		Delta   float32       `envconfig:"optional"`
		DeltaV  float64       `envconfig:"optional"`
		Slice   []string      `envconfig:"optional"`
		Struct  struct {
			A string
			B int
		} `envconfig:"optional"`
	}

	os.Setenv("NAME", "")
	os.Setenv("FLAG", "")
	os.Setenv("TIMEOUT", "")
	os.Setenv("PORT", "")
	os.Setenv("PORT2", "")
	os.Setenv("DELTA", "")
	os.Setenv("DELTAV", "")
	os.Setenv("SLICE", "")
	os.Setenv("STRUCT", "")

	err := envconfig.Init(&conf)
	require.Nil(t, err)
	require.Equal(t, "", conf.Name)
}

func TestParseSkippableConfig(t *testing.T) {
	var conf struct {
		Flag bool `envconfig:"-"`
	}

	os.Setenv("FLAG", "true")

	err := envconfig.Init(&conf)
	require.Nil(t, err)
	require.Equal(t, false, conf.Flag)
}

func TestParseCustomNameConfig(t *testing.T) {
	var conf struct {
		Name string `envconfig:"customName"`
	}

	os.Setenv("customName", "foobar")

	err := envconfig.Init(&conf)
	require.Nil(t, err)
	require.Equal(t, "foobar", conf.Name)
}

func TestParseOptionalStruct(t *testing.T) {
	var conf struct {
		Master struct {
			Name string
		} `envconfig:"optional"`
	}

	os.Setenv("MASTER_NAME", "")

	err := envconfig.Init(&conf)
	require.Nil(t, err)
	require.Equal(t, "", conf.Master.Name)
}

func TestParsePrefixedStruct(t *testing.T) {
	var conf struct {
		Name string
	}

	os.Setenv("NAME", "")
	os.Setenv("FOO_NAME", "")

	os.Setenv("NAME", "bad")
	err := envconfig.InitWithPrefix(&conf, "FOO")
	require.NotNil(t, err)

	os.Setenv("FOO_NAME", "good")
	err = envconfig.InitWithPrefix(&conf, "FOO")
	require.Nil(t, err)
	require.Equal(t, "good", conf.Name)
}

func TestUnexportedField(t *testing.T) {
	var conf struct {
		name string
	}

	os.Setenv("NAME", "foobar")

	err := envconfig.Init(&conf)
	require.Equal(t, envconfig.ErrUnexportedField, err)
}

type sliceWithUnmarshaler []int

func (sl *sliceWithUnmarshaler) Unmarshal(s string) error {
	tokens := strings.Split(s, ".")
	for _, tok := range tokens {
		tmp, err := strconv.Atoi(tok)
		if err != nil {
			return err
		}

		*sl = append(*sl, tmp)
	}

	return nil
}

func TestSliceTypeWithUnmarshaler(t *testing.T) {
	var conf struct {
		Data sliceWithUnmarshaler
	}

	os.Setenv("DATA", "1.2.3")

	err := envconfig.Init(&conf)
	require.Nil(t, err)
	require.Equal(t, 3, len(conf.Data))
	require.Equal(t, 1, conf.Data[0])
	require.Equal(t, 2, conf.Data[1])
	require.Equal(t, 3, conf.Data[2])
}

func TestParseDefaultVal(t *testing.T) {
	var conf struct {
		MySQL struct {
			Master struct {
				Address string `envconfig:"default=localhost"`
				Port    int    `envconfig:"default=3306"`
			}
			Timeout time.Duration `envconfig:"default=1m,myTimeout"`
		}
	}

	err := envconfig.Init(&conf)
	require.Nil(t, err)
	require.Equal(t, "localhost", conf.MySQL.Master.Address)
	require.Equal(t, 3306, conf.MySQL.Master.Port)
	require.Equal(t, time.Minute*1, conf.MySQL.Timeout)

	os.Setenv("myTimeout", "2m")

	err = envconfig.Init(&conf)
	require.Nil(t, err)
	require.Equal(t, "localhost", conf.MySQL.Master.Address)
	require.Equal(t, 3306, conf.MySQL.Master.Port)
	require.Equal(t, time.Minute*2, conf.MySQL.Timeout)
}

func TestInitNotAPointer(t *testing.T) {
	err := envconfig.Init("foobar")
	require.Equal(t, envconfig.ErrNotAPointer, err)
}

func TestInitPointerToAPointer(t *testing.T) {
	type Conf struct {
		Name string
	}
	var tmp *Conf

	os.Setenv("NAME", "foobar")

	err := envconfig.Init(&tmp)
	require.Nil(t, err)
	require.Equal(t, "foobar", tmp.Name)
}

func TestInitInvalidValueKind(t *testing.T) {
	sl := []string{"foo", "bar"}
	err := envconfig.Init(&sl)
	require.Equal(t, envconfig.ErrInvalidValueKind, err)
}

func TestInvalidFieldValueKind(t *testing.T) {
	var conf struct {
		Foo interface{}
	}

	os.Setenv("FOO", "lalala")

	err := envconfig.Init(&conf)
	require.Equal(t, "envconfig: kind interface not supported", err.Error())
}

func TestInvalidSliceElementValueKind(t *testing.T) {
	var conf struct {
		Foo []interface{}
	}

	os.Setenv("FOO", "lalala")

	err := envconfig.Init(&conf)
	require.Equal(t, "envconfig: kind interface not supported", err.Error())
}

func TestParseEmptyTag(t *testing.T) {
	var conf struct {
		Name string `envconfig:""`
	}

	os.Setenv("NAME", "foobar")

	err := envconfig.Init(&conf)
	require.Nil(t, err)
	require.Equal(t, "foobar", conf.Name)
}
