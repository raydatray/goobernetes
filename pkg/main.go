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
	"github.com/raydatray/goobernetes/pkg/utils"
	"github.com/spf13/cobra"
)

func main() {
	var config utils.Config

	rootCmd := &cobra.Command{
		Use:   "goobernetes",
		Short: "a simple load balancer implementation",
	}

	lbCmd := &cobra.Command{
		Use:   "lb",
		Short: "start a load balancer instance",
		Run: func(cmd *cobra.Command, args []string) {
			lb := loadbalancer.NewRoundRobinLoadBalancer()

			server1, _ := loadbalancer.NewServerInstance("mcschool", "192.0.0.1", 8081, 5, 0)
			server2, _ := loadbalancer.NewServerInstance("g1-home-router", "192.0.0.2", 8082, 5, 0)
			server3, _ := loadbalancer.NewServerInstance("herroshima", "192.0.0.3", 8083, 5, 0)

			defaultBackends := []*loadbalancer.ServerInstance{
				server1,
				server2,
				server3,
			}

			for _, server := range defaultBackends {
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
			srv := servlets.NewBackendServer(config)

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
		cmd.Flags().IntVarP(&config.Connections, "connections", "c", 100, "maximum number of concurrent connections")
		cmd.Flags().IntVarP(&config.ConnectionPoolSize, "connection-pool-size", "s", 10, "size of the connection pool")
	}

	rootCmd.AddCommand(lbCmd, backendCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
