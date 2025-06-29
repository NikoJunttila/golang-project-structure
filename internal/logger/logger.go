package logger
import (
	"io"
	"os"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"gopkg.in/natefinch/lumberjack.v2"
)

func Error(err error, message string){
	log.Error().Err(err).Msg(message)
}
func Warn(message string){
	log.Warn().Msg(message)
}
func Info(message string){
	log.Info().Msg(message)
}
func Debug(message string){
	log.Debug().Msg(message)
}
func Fatal(err error,message string){
	log.Fatal().Err(err).Msg(message)
}

func LoggerSetup() {
	// Create the logs directory if it doesn't exist
	if _, err := os.Stat("logs"); os.IsNotExist(err) {
		os.Mkdir("logs", 0755)
	}
	// Set up lumberjack logger for file rotation
	logFile := &lumberjack.Logger{
		Filename:   "logs/app.log", // Log file name
		MaxSize:    20,             // Max size in MB before rotation
		MaxBackups: 5,              // Max number of old log files to keep
		MaxAge:     28,             // Max age in days to keep a log file
		Compress:   true,           // Compress rotated files
	}
  zerolog.SetGlobalLevel(zerolog.DebugLevel)
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	// Optional: Also log to console
	consoleWriter := zerolog.ConsoleWriter{
		Out:        os.Stdout,
		TimeFormat: "02-01-2006 15:04:05",
	}
	// Combine both outputs
	multi := io.MultiWriter(logFile, consoleWriter)

	// Set up zerolog with combined output, timestamp and caller info
	log.Logger = zerolog.New(multi).With().Timestamp().Caller().Logger()
	Info("Global logger initialized")
}
