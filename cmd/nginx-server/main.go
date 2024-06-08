package main

import (
	"flag"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"syscall"
	"time"

	// Packages
	server "github.com/mutablelogic/go-server"
	ctx "github.com/mutablelogic/go-server/pkg/context"
	auth "github.com/mutablelogic/go-server/pkg/handler/auth"
	certmanager "github.com/mutablelogic/go-server/pkg/handler/certmanager"
	certstore "github.com/mutablelogic/go-server/pkg/handler/certmanager/certstore"
	ldap "github.com/mutablelogic/go-server/pkg/handler/ldap"
	logger "github.com/mutablelogic/go-server/pkg/handler/logger"
	nginx "github.com/mutablelogic/go-server/pkg/handler/nginx"
	router "github.com/mutablelogic/go-server/pkg/handler/router"
	tokenjar "github.com/mutablelogic/go-server/pkg/handler/tokenjar"
	httpserver "github.com/mutablelogic/go-server/pkg/httpserver"
	provider "github.com/mutablelogic/go-server/pkg/provider"
)

var (
	binary        = flag.String("path", "nginx", "Path to nginx binary")
	group         = flag.String("group", "", "Group to run unix socket as")
	ldap_password = flag.String("ldap-password", "", "LDAP admin password")
)

/* command to test the nginx package */
/* will run the nginx server and provide an nginx api for reloading,
   testing, etc through FastCGI. The config and run paths are a bit
   screwed up and will need to be fixed.
*/
func main() {
	flag.Parse()

	// Create context which cancels on interrupt
	ctx := ctx.ContextForSignal(os.Interrupt, syscall.SIGQUIT)

	// Logger
	logger, err := logger.Config{Flags: []string{"default", "prefix"}}.New()
	if err != nil {
		log.Fatal("logger: ", err)
	}

	// Nginx handler
	n, err := nginx.Config{BinaryPath: *binary}.New()
	if err != nil {
		log.Fatal("nginx: ", err)
	}

	// Token Jar
	jar, err := tokenjar.Config{
		DataPath:      n.(nginx.Nginx).Config(),
		WriteInterval: 30 * time.Second,
	}.New()
	if err != nil {
		log.Fatal("tokenkar: ", err)
	}

	// Auth handler
	auth, err := auth.Config{
		TokenJar:   jar.(auth.TokenJar),
		TokenBytes: 8,
		Bearer:     true, // Use bearer token in requests for authorization
	}.New()
	if err != nil {
		log.Fatal("auth: ", err)
	}

	// Cert Storage
	certstore, err := certstore.Config{
		DataPath: filepath.Join(n.(nginx.Nginx).Config(), "cert"),
		Group:    *group,
	}.New()
	if err != nil {
		log.Fatal("certstore: ", err)
	}
	certmanager, err := certmanager.Config{
		CertStorage: certstore.(certmanager.CertStorage),
	}.New()
	if err != nil {
		log.Fatal("certmanager: ", err)
	}

	// LDAP
	ldap, err := ldap.Config{
		URL:      "ldap://admin@cm1.local/",
		DN:       "dc=mutablelogic,dc=com",
		Password: *ldap_password,
	}.New()
	if err != nil {
		log.Fatal("ldap: ", err)
	}

	// Location of the FCGI unix socket
	socket := filepath.Join(n.(nginx.Nginx).Config(), "run/go-server.sock")

	// Router
	router, err := router.Config{
		Services: router.ServiceConfig{
			"nginx": { // /api/nginx/...
				Service: n.(server.ServiceEndpoints),
				Middleware: []server.Middleware{
					logger.(server.Middleware),
					auth.(server.Middleware),
				},
			},
			"auth": { // /api/auth/...
				Service: auth.(server.ServiceEndpoints),
				Middleware: []server.Middleware{
					logger.(server.Middleware),
					auth.(server.Middleware),
				},
			},
			"cert": { // /api/cert/...
				Service: certmanager.(server.ServiceEndpoints),
				Middleware: []server.Middleware{
					logger.(server.Middleware),
					auth.(server.Middleware),
				},
			},
			"ldap": { // /api/ldap/...
				Service: ldap.(server.ServiceEndpoints),
				Middleware: []server.Middleware{
					logger.(server.Middleware),
					auth.(server.Middleware),
				},
			},
		},
	}.New()
	if err != nil {
		log.Fatal("router: ", err)
	}

	// HTTP Server
	httpserver, err := httpserver.Config{
		Listen: socket,
		Group:  *group,
		Router: router.(http.Handler),
	}.New()
	if err != nil {
		log.Fatal("httpserver: ", err)
	}

	// Run until we receive an interrupt
	provider := provider.NewProvider(logger, n, jar, auth, certstore, certmanager, ldap, router, httpserver)
	provider.Print(ctx, "Press CTRL+C to exit")
	if err := provider.Run(ctx); err != nil {
		log.Fatal(err)
	}
}
