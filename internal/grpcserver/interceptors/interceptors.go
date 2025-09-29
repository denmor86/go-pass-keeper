package grpcserver

import (
	"context"
	"fmt"
	"go-pass-keeper/pkg/logger"
	"go-pass-keeper/pkg/usercontext"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/google/uuid"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/auth"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/recovery"
)

// tokenHandler интефрейс для работы с токеном
type tokenHandler interface {
	// DecodeUserId - извлечение ID пользователя из токена
	DecodeUserId(token string) (string, error)
}

// MakeAuthFunc - метод создания функции авторизации для перехватчика
func MakeAuthFunc(handler tokenHandler) auth.AuthFunc {
	return func(ctx context.Context) (context.Context, error) {
		jwt, err := auth.AuthFromMD(ctx, "bearer")
		if err != nil {
			return nil, err
		}

		uid, err := handler.DecodeUserId(jwt)
		if err != nil {
			return nil, status.Errorf(codes.Unauthenticated, "invalid auth token: %v", err)
		}

		u, err := uuid.Parse(uid)
		if err != nil {
			return nil, status.Errorf(codes.Unauthenticated, "invalid auth token: %v", err)
		}

		// создаем контекст, и добавляем в него ID пользователя (чтобы отвязать обработчик от парсинга cookie)
		ctx = usercontext.SetUserId(ctx, u)
		return ctx, nil
	}
}

// InterceptorLogger - метод перехватчик логирования в GRPC
func InterceptorLogger(l *zap.Logger) logging.Logger {
	return logging.LoggerFunc(func(ctx context.Context, lvl logging.Level, msg string, fields ...any) {
		f := make([]zap.Field, 0, len(fields)/2)

		for i := 0; i < len(fields); i += 2 {
			key := fields[i]
			value := fields[i+1]

			switch v := value.(type) {
			case string:
				f = append(f, zap.String(key.(string), v))
			case int:
				f = append(f, zap.Int(key.(string), v))
			case bool:
				f = append(f, zap.Bool(key.(string), v))
			default:
				f = append(f, zap.Any(key.(string), v))
			}
		}

		logger := l.WithOptions(zap.AddCallerSkip(1)).With(f...)

		switch lvl {
		case logging.LevelDebug:
			logger.Debug(msg)
		case logging.LevelInfo:
			logger.Info(msg)
		case logging.LevelWarn:
			logger.Warn(msg)
		case logging.LevelError:
			logger.Error(msg)
		default:
			panic(fmt.Sprintf("unknown level %v", lvl))
		}
	})
}

// CreateUnaryInterceptors - метод для создания перехватчиков обычных запросов
func CreateUnaryInterceptors(handler tokenHandler) []grpc.UnaryServerInterceptor {
	return []grpc.UnaryServerInterceptor{

		logging.UnaryServerInterceptor(InterceptorLogger(logger.Get().Desugar())),
		recovery.UnaryServerInterceptor(),
		auth.UnaryServerInterceptor(MakeAuthFunc(handler)),
	}
}

// CreateStreamInterceptors - метод для создания перехватчиков потоковых запросов
func CreateStreamInterceptors(handler tokenHandler) []grpc.StreamServerInterceptor {
	return []grpc.StreamServerInterceptor{
		auth.StreamServerInterceptor(MakeAuthFunc(handler)),
	}
}
