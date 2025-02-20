package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/raydatray/goobernetes/pkg/loadbalancer"
	"github.com/raydatray/goobernetes/pkg/router"
	"github.com/raydatray/goobernetes/pkg/servlets"
	"github.com/spf13/cobra"
)

type Config struct {
	Port int
}

func main() {
	var config Config

	rootCmd := &cobra.Command{
		Use:   "goobernetes",
		Short: "a simple load balancer implementation",
	}

	lbCmd := &cobra.Command{
		Use:   "lb",
		Short: "start a load balancer instance",
		Run: func(cmd *cobra.Command, args []string) {
			lb := loadbalancer.NewRoundRobinLoadBalancer()

			servers := []struct {
				id      string
				host    string
				port    int
				maxConn int
			}{
				{"mcschool", "192.168.23.145", 8081, 5},
				{"g1 home router", "10.234.67.89", 8082, 5},
				{"typeracer", "172.16.92.201", 8083, 5},
			}

			for _, server := range servers {
				server, err := loadbalancer.NewServerInstance(
					server.id,
					server.host,
					server.port,
					server.maxConn,
				)

				if err != nil {
					log.Printf("failed to create server: %v", err)
				}

				if err := lb.AddServer(server); err != nil {
					log.Printf("failed to add server: %v", err)
				}
			}

			r := router.NewRouter(lb)
			srv := servlets.NewHttpServer(r, config.Port)

			sigChan := make(chan os.Signal, 1)
			signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

			errChan := make(chan error, 1)
			go func() {
				errChan <- srv.Start()
			}()

			select {
			case err := <-errChan:
				if err != nil {
					log.Fatalf("server error: %v", err)
				}
			case sig := <-sigChan:
				log.Printf("received signal: %v", sig)
				if err := srv.Stop(); err != nil {
					log.Printf("error during shutdown: %v", err)
				}
			}
		},
	}

	backendCmd := &cobra.Command{
		Use:   "backend",
		Short: "start a mock backend server instance",
		Run: func(cmd *cobra.Command, args []string) {
			srv := servlets.NewBackendServer(config.Port)

			sigChan := make(chan os.Signal, 1)
			signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

			errChan := make(chan error, 1)
			go func() {
				errChan <- srv.Start()
			}()

			select {
			case err := <-errChan:
				if err != nil {
					log.Fatalf("server error: %v", err)
				}
			case sig := <-sigChan:
				log.Printf("received signal: %v", sig)
				if err := srv.Stop(); err != nil {
					log.Printf("error during shutdown: %v", err)
				}
			}
		},
	}

	for _, cmd := range []*cobra.Command{lbCmd, backendCmd} {
		cmd.Flags().IntVarP(&config.Port, "port", "p", 8080, "port to run the server on")
	}

	rootCmd.AddCommand(lbCmd, backendCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
