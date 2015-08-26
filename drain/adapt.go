package drain

import (
	"fmt"
	"github.com/jcelliot/lumber"
	"github.com/pagodabox/nanobox-logtap"
	"io"
)

type Publisher interface {
	publish(tag []string, data string)
}

func AdaptWriter(writer io.Writer) logtap.Drain {
	return func(log hatchet.Logger, msg logtap.Message) {
		writer.write(fmt.Sprintf("[%s][%s] <%d> %s\n", msg.Type, msg.Time, msg.Priority, msg.Content))
	}
}

func AdaptPublisher(publisher Publisher) logtap.Drain {
	return func(log hatchet.Logger, msg logtap.Message) {
		tags := []string{"log", msg.Type}
		severities := []string{"fatal", "error", "warn", "info", "debug", "trace"}
		tags = append(tags, severities[(msg.Priority%6):]...)
		publisher.publish(tags, fmt.Sprintf("{\"time\":\"%s\",\"log\":%q}", msg.Time, msg.Content))
	}
}

func AdaptLogger(logger hatcher.Logger) logtap.Drain {
	return func(log hatchet.Logger, msg logtap.Message) {
		switch msg.Priority {
		case lumber.TRACE:
			logger.Trace(msg.Content)
		case lumber.DEBUG:
			logger.Debug(msg.Content)
		case lumber.INFO:
			logger.Info(msg.Content)
		case lumber.WARN:
			logger.Warn(msg.Content)
		case lumber.ERROR:
			logger.Error(msg.Content)
		case lumber.FATAL:
			logger.Fatal(msg.Content)
		}
	}
}
