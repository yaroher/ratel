package atlas

import (
	"ariga.io/atlas/sql/migrate"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type migrationLogger struct {
	lg *zap.Logger
}

func NewMigrationLogger(logger2 *zap.Logger) migrate.Logger {
	return migrationLogger{lg: logger2}
}
func (m migrationLogger) listFilesToLog(files []migrate.File) []zap.Field {
	ret := make([]zap.Field, 0, len(files))
	for _, f := range files {
		ret = append(ret, zap.String("file", f.Name()))
	}
	return ret
}

func (m migrationLogger) Log(entry migrate.LogEntry) {
	switch e := entry.(type) {
	case migrate.LogExecution:
		m.lg.Debug(
			"migration execution",
			append(m.listFilesToLog(e.Files),
				zap.String("from", e.From),
				zap.String("to", e.To),
			)...,
		)
	case migrate.LogStmt:
		m.lg.Debug(
			"migration statement",
			zap.String("sql", e.SQL),
		)
	case migrate.LogError:
		m.lg.Error(
			"migration error",
			zap.String("sql", e.SQL),
			zap.Error(e.Error),
		)
	case migrate.LogChecks:
		m.lg.Debug(
			"migration checks",
			zap.String("name", e.Name),
			zap.Array("stmts", zapcore.ArrayMarshalerFunc(func(encoder zapcore.ArrayEncoder) error {
				for _, str := range e.Stmts {
					encoder.AppendString(str)
				}
				return nil
			})),
		)
	case migrate.LogCheck:
		m.lg.Debug(
			"migration check",
			zap.String("stmt", e.Stmt),
			zap.Error(e.Error),
		)
	case migrate.LogChecksDone:
		m.lg.Debug(
			"migration checks done",
			zap.Error(e.Error),
		)
	}
}
