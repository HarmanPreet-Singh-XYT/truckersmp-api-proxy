package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

const (
	TruckersMPAPIBase = "https://api.truckersmp.com/v2"
	ServerPort        = ":4004"
)

type ProxyServer struct {
	client *http.Client
}

func NewProxyServer() *ProxyServer {
	return &ProxyServer{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// Generic proxy handler that forwards requests to TruckersMP API
func (p *ProxyServer) proxyRequest(c *gin.Context, endpoint string) {
	// Build the full URL
	url := TruckersMPAPIBase + endpoint

	// Create the request
	req, err := http.NewRequest(c.Request.Method, url, nil)
	req.Header.Set("User-Agent", "PostmanRuntime/7.36.1")
	req.Header.Set("Accept", "application/json")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   true,
			"message": "Failed to create request",
		})
		return
	}

	// Copy headers from original request
	for key, values := range c.Request.Header {
		for _, value := range values {
			req.Header.Add(key, value)
		}
	}

	// Make the request
	resp, err := p.client.Do(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   true,
			"message": "Failed to fetch data from TruckersMP API",
		})
		return
	}
	defer resp.Body.Close()

	// Read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   true,
			"message": "Failed to read response",
		})
		return
	}

	// Copy response headers
	for key, values := range resp.Header {
		for _, value := range values {
			c.Header(key, value)
		}
	}

	// Set the status code and return the response
	c.Status(resp.StatusCode)

	// Try to parse as JSON and return formatted, otherwise return raw
	var jsonResponse interface{}
	if json.Unmarshal(body, &jsonResponse) == nil {
		c.JSON(resp.StatusCode, jsonResponse)
	} else {
		c.Data(resp.StatusCode, resp.Header.Get("Content-Type"), body)
	}
}

// Validation helpers
func validateID(c *gin.Context, paramName string) (int, bool) {
	idStr := c.Param(paramName)
	id, err := strconv.Atoi(idStr)
	if err != nil || id < 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   true,
			"message": fmt.Sprintf("Invalid %s parameter", paramName),
		})
		return 0, false
	}
	return id, true
}

func (p *ProxyServer) setupRoutes() *gin.Engine {
	r := gin.Default()

	// Get allowed origin from environment variable
	allowedOrigin := os.Getenv("ORIGIN")
	if allowedOrigin == "" {
		allowedOrigin = "*" // fallback
	}

	// Add CORS middleware
	r.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", allowedOrigin)
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})

	// Health check endpoint
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":    "healthy",
			"service":   "TruckersMP API Proxy",
			"timestamp": time.Now().Unix(),
		})
	})

	// Player endpoints
	r.GET("/player/:id", func(c *gin.Context) {
		if id, valid := validateID(c, "id"); valid {
			p.proxyRequest(c, fmt.Sprintf("/player/%d", id))
		}
	})

	// Bans endpoints
	r.GET("/bans/:id", func(c *gin.Context) {
		if id, valid := validateID(c, "id"); valid {
			p.proxyRequest(c, fmt.Sprintf("/bans/%d", id))
		}
	})

	// Servers endpoint
	r.GET("/servers", func(c *gin.Context) {
		p.proxyRequest(c, "/servers")
	})

	// Game time endpoint
	r.GET("/game_time", func(c *gin.Context) {
		p.proxyRequest(c, "/game_time")
	})

	// Events endpoints
	r.GET("/events", func(c *gin.Context) {
		p.proxyRequest(c, "/events")
	})

	r.GET("/events/:id", func(c *gin.Context) {
		if id, valid := validateID(c, "id"); valid {
			p.proxyRequest(c, fmt.Sprintf("/events/%d", id))
		}
	})

	r.GET("/events/user/:id", func(c *gin.Context) {
		if id, valid := validateID(c, "id"); valid {
			p.proxyRequest(c, fmt.Sprintf("/events/user/%d", id))
		}
	})

	// VTC endpoints
	r.GET("/vtc", func(c *gin.Context) {
		p.proxyRequest(c, "/vtc")
	})

	r.GET("/vtc/:id", func(c *gin.Context) {
		if id, valid := validateID(c, "id"); valid {
			p.proxyRequest(c, fmt.Sprintf("/vtc/%d", id))
		}
	})

	r.GET("/vtc/:id/news", func(c *gin.Context) {
		if id, valid := validateID(c, "id"); valid {
			p.proxyRequest(c, fmt.Sprintf("/vtc/%d/news", id))
		}
	})

	r.GET("/vtc/:id/news/:news_id", func(c *gin.Context) {
		if id, valid := validateID(c, "id"); valid {
			if newsID, valid := validateID(c, "news_id"); valid {
				p.proxyRequest(c, fmt.Sprintf("/vtc/%d/news/%d", id, newsID))
			}
		}
	})

	r.GET("/vtc/:id/roles", func(c *gin.Context) {
		if id, valid := validateID(c, "id"); valid {
			p.proxyRequest(c, fmt.Sprintf("/vtc/%d/roles", id))
		}
	})

	r.GET("/vtc/:id/role/:role_id", func(c *gin.Context) {
		if id, valid := validateID(c, "id"); valid {
			if roleID, valid := validateID(c, "role_id"); valid {
				p.proxyRequest(c, fmt.Sprintf("/vtc/%d/role/%d", id, roleID))
			}
		}
	})

	r.GET("/vtc/:id/members", func(c *gin.Context) {
		if id, valid := validateID(c, "id"); valid {
			p.proxyRequest(c, fmt.Sprintf("/vtc/%d/members", id))
		}
	})

	r.GET("/vtc/:id/member/:member_id", func(c *gin.Context) {
		if id, valid := validateID(c, "id"); valid {
			if memberID, valid := validateID(c, "member_id"); valid {
				p.proxyRequest(c, fmt.Sprintf("/vtc/%d/member/%d", id, memberID))
			}
		}
	})

	r.GET("/vtc/:id/events", func(c *gin.Context) {
		if id, valid := validateID(c, "id"); valid {
			p.proxyRequest(c, fmt.Sprintf("/vtc/%d/events", id))
		}
	})

	r.GET("/vtc/:id/events/:event_id", func(c *gin.Context) {
		if id, valid := validateID(c, "id"); valid {
			if eventID, valid := validateID(c, "event_id"); valid {
				p.proxyRequest(c, fmt.Sprintf("/vtc/%d/events/%d", id, eventID))
			}
		}
	})

	r.GET("/vtc/:id/events/attending", func(c *gin.Context) {
		if id, valid := validateID(c, "id"); valid {
			p.proxyRequest(c, fmt.Sprintf("/vtc/%d/events/attending", id))
		}
	})

	r.GET("/vtc/:id/partners", func(c *gin.Context) {
		if id, valid := validateID(c, "id"); valid {
			p.proxyRequest(c, fmt.Sprintf("/vtc/%d/partners", id))
		}
	})

	// Version endpoint
	r.GET("/version", func(c *gin.Context) {
		p.proxyRequest(c, "/version")
	})

	// Rules endpoint
	r.GET("/rules", func(c *gin.Context) {
		p.proxyRequest(c, "/rules")
	})

	// Catch-all for any other routes
	r.NoRoute(func(c *gin.Context) {
		c.JSON(http.StatusNotFound, gin.H{
			"error":   true,
			"message": "Endpoint not found",
		})
	})

	return r
}

func main() {
	// Set Gin to release mode in production
	gin.SetMode(gin.ReleaseMode)

	// Create proxy server
	proxy := NewProxyServer()

	// Setup routes
	router := proxy.setupRoutes()

	// Start server
	log.Printf("Starting TruckersMP API Proxy Server on port %s", ServerPort)
	log.Printf("Proxying requests to: %s", TruckersMPAPIBase)

	if err := router.Run(ServerPort); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
