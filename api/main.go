package main

import (
	"context"
	"flag"
	"log"
	"net/http"

	"git.neds.sh/matty/entain/api/proto/sports_pb"

	"git.neds.sh/matty/entain/api/proto/racing"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
)

var (
	apiEndpoint        = flag.String("api-endpoint", "localhost:8000", "API endpoint")
	racesGrpcEndpoint  = flag.String("races-grpc-endpoint", "localhost:9000", "Races gRPC server endpoint")
	sportsGrpcEndpoint = flag.String("sports-grpc-endpoint", "localhost:9001", "Sports gRPC server endpoint")
)

func main() {
	flag.Parse()

	if err := run(); err != nil {
		log.Printf("failed running api server: %s\n", err)
	}
}

func run() error {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	mux := runtime.NewServeMux()

	if err := sports_pb.RegisterSportsHandlerFromEndpoint(
		ctx,
		mux,
		*sportsGrpcEndpoint,
		[]grpc.DialOption{grpc.WithInsecure()},
	); err != nil {
		return err
	}

	if err := racing.RegisterRacingHandlerFromEndpoint(
		ctx,
		mux,
		*racesGrpcEndpoint,
		[]grpc.DialOption{grpc.WithInsecure()},
	); err != nil {
		return err
	}

	log.Printf("API server listening on: %s\n", *apiEndpoint)

	return http.ListenAndServe(*apiEndpoint, mux)
}
