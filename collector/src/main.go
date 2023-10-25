// Copyright The OpenTelemetry Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/open-telemetry/opentelemetry-lambda/collector/internal/lifecycle"
)

var (
	// Version variable will be replaced at link time after `make` has been run.
	Version = "latest"

	// GitHash variable will be replaced at link time after `make` has been run.
	GitHash = "<NOT PROPERLY GENERATED>"
)

func main() {
	versionFlag := flag.Bool("v", false, "prints version information")
	flag.Parse()
	if *versionFlag {
		fmt.Println(Version)
		return
	}

	logger := initLogger()
	logger.Info("Launching OpenTelemetry Lambda extension", zap.String("version", Version))

	// Check for deprecated usage of SUMOLOGIC_HTTP_TRACES_URL environment variable
	sumoLogicOtlpHttpEndpoint, ok := os.LookupEnv("SUMOLOGIC_HTTP_TRACES_ENDPOINT_URL")
	if ok {
		logger.Warn("SUMOLOGIC_HTTP_TRACES_ENDPOINT_URL is deprecated. Use SUMO_OTLP_HTTP_ENDPOINT_URL environment variable instead. Please see OTLP/http endpoint generation guide https://help.sumologic.com/docs/send-data/hosted-collectors/http-source/otlp/.")
		os.Setenv("SUMO_OTLP_HTTP_ENDPOINT_URL", sumoLogicOtlpHttpEndpoint)
	}

	ctx, lm := lifecycle.NewManager(context.Background(), logger, Version)

	// Will block until shutdown event is received or cancelled via the context.
	logger.Info("done", zap.Error(lm.Run(ctx)))
}

func initLogger() *zap.Logger {
	lvl := zap.NewAtomicLevelAt(zapcore.InfoLevel)

	envLvl := os.Getenv("OPENTELEMETRY_EXTENSION_LOG_LEVEL")
	userLvl, err := zap.ParseAtomicLevel(envLvl)
	if err == nil {
		lvl = userLvl
	}

	l := zap.New(zapcore.NewCore(zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig()), os.Stdout, lvl))

	if err != nil && envLvl != "" {
		l.Warn("unable to parse log level from environment", zap.Error(err))
	}

	return l
}
