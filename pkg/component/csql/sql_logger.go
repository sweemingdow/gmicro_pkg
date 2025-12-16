package csql

import "github.com/sweemingdow/gmicro_pkg/pkg/mylog"

const moduleLogger = "sqlLogger"

type sqlLogger struct{}

func (s *sqlLogger) Event(eventName string) {
	lg := mylog.GetLogger(moduleLogger)
	lg.Debug().Msgf("[Event]:%s", eventName)
}

func (s *sqlLogger) EventKv(eventName string, kvs map[string]string) {
	lg := mylog.GetLogger(moduleLogger)
	lg.Debug().Msgf("[EventKv]:%s, kvs:%v", eventName, kvs)
}

func (s *sqlLogger) EventErr(eventName string, err error) error {
	lg := mylog.GetLogger(moduleLogger)
	lg.Error().Stack().Err(err).Msgf("[EventErr]:%s", eventName)

	return err
}

func (s *sqlLogger) EventErrKv(eventName string, err error, kvs map[string]string) error {
	lg := mylog.GetLogger(moduleLogger)
	lg.Error().Stack().Err(err).Msgf("[EventErrKv]:%s, kvs:%v", eventName, kvs)
	return err
}

func (s *sqlLogger) Timing(eventName string, nanoseconds int64) {
	lg := mylog.GetLogger(moduleLogger)
	lg.Debug().Msgf("[Timing]:%s, duration:%dns", eventName, nanoseconds)
}

func (s *sqlLogger) TimingKv(eventName string, nanoseconds int64, kvs map[string]string) {
	lg := mylog.GetLogger(moduleLogger)
	lg.Debug().Msgf("[TimingKv]:%s, kvs:%v, duration:%dns,", eventName, kvs, nanoseconds)
}
