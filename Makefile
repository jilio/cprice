build:
	go build \
		-ldflags "-X main.SentryDSN=$(cat sentry_dsn)" \
		-o ./dist/CPrice.app/Contents/MacOS/cprice