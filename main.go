package main

import (
	"fmt"
	"net/http"
	"github.com/gorilla/handlers"
	"log"
	"github.com/garyburd/redigo/redis"
	"os"
	"time"
	"os/signal"
	"syscall"
	"strings"
)

var (
	Pool *redis.Pool
	appSettings AppSettings
)

func init() {
	// TODO: proper config file
	appSettings = AppSettings{
		RedisHost: "[ec2-52-200-201-70.compute-1.amazonaws.com]:6379",
		RedisPassword: "a71f97758647d4aa5b093943db3b16c8656dfdb7ef1289fc5b7b37c1e2895e63",
		Env: "local",
	}


	redisHost := appSettings.RedisHost
	password := appSettings.RedisPassword
	Pool = newPool(redisHost, password)
	cleanupHook()
	fmt.Println("Connected to Redis")
	fmt.Println("Initialization Complete")
}

func main() {
	fmt.Println("Server Starting")

	r := MakeHttpHandler(Pool)

	// Cors Options
	headersOk := handlers.AllowedHeaders([]string{"X-Requested-With", "Content-Type", "Access-Control-Allow-Origin: *"})
	originsOk := handlers.AllowedOrigins([]string{"*"})
	methodsOk := handlers.AllowedMethods([]string{"GET", "HEAD", "POST", "PUT", "OPTIONS"})

	//  Start HTTPS
	if strings.ToLower(appSettings.Env) == "prod"{
		go func() {
			err_https := http.ListenAndServeTLS(":443", "fullchain.pem", "privkey.pem", handlers.CORS(originsOk, headersOk, methodsOk)(r))
			if err_https != nil {
				log.Fatal("Web server (HTTPS): ", err_https)
			}
		}()
	}

	//  Start HTTP
	err_http := http.ListenAndServe(":80", handlers.CORS(originsOk, headersOk, methodsOk)(r))
	if err_http != nil {
		log.Fatal("Web server (HTTP): ", err_http)
	}

}


/**
*  Initialize a new Redis pool
 */
func newPool(server string, password string) *redis.Pool {
	return &redis.Pool{

		MaxIdle:     3,
		IdleTimeout: 240 * time.Second,

		Dial: func() (redis.Conn, error) {
			c, err := redis.Dial("tcp", server)
			if err != nil {
				return nil, err
			}
			_, err = c.Do("AUTH", password)
			if err != nil {
				return nil, err
			}
			return c, err
		},

		TestOnBorrow: func(c redis.Conn, t time.Time) error {
			_, err := c.Do("PING")
			return err
		},
	}
}

func cleanupHook() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	signal.Notify(c, syscall.SIGTERM)
	signal.Notify(c, syscall.SIGKILL)
	go func() {
		<-c
		Pool.Close()
		os.Exit(0)
	}()
}