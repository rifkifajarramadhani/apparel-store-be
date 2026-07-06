package router

import (
	"log/slog"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/limiter"
	"github.com/google/uuid"
	"github.com/rifkifajarramadhani/golang-clean-architecture/internal/adapter/http/handler"
	"github.com/rifkifajarramadhani/golang-clean-architecture/internal/adapter/http/middleware"
	"github.com/rifkifajarramadhani/golang-clean-architecture/internal/storage"
)

func Setup(app *fiber.App, users handler.UserService, auth handler.AuthService, merchandising handler.MerchandisingServices, orders handler.OrderService, tokens middleware.AccessTokenValidator, uploader storage.ImageUploader, logger *slog.Logger, storefrontURL ...string) {
	// Liveness/readiness probes for the container healthcheck and zero-downtime
	// rollouts. Registered before the logging middleware so the frequent polls
	// don't flood the request log, and outside /api so they need no auth.
	health := func(c fiber.Ctx) error { return c.SendStatus(fiber.StatusOK) }
	app.Get("/health/live", health)
	app.Get("/health/ready", health)

	app.Use(func(c fiber.Ctx) error {
		requestID := c.Get("X-Request-ID")
		if requestID == "" {
			requestID = uuid.NewString()
		}

		c.Set("X-Request-ID", requestID)

		started := time.Now()
		err := c.Next()
		logger.InfoContext(c.Context(), "http request",
			"request_id", requestID, "method", c.Method(), "path", c.Path(),
			"status", c.Response().StatusCode(), "duration", time.Since(started),
		)
		return err
	})

	api := app.Group("/api")
	authGroup := api.Group("/auth", limiter.New(limiter.Config{
		Max: 20, Expiration: time.Minute,
		LimitReached: func(c fiber.Ctx) error {
			return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{"error": "too many requests"})
		},
	}))
	authHandler := handler.NewAuthHandler(auth, logger, storefrontURL...)
	authGroup.Post("/register", authHandler.Register)
	authGroup.Post("/login", authHandler.Login)
	authGroup.Post("/refresh", authHandler.Refresh)
	authGroup.Post("/logout", authHandler.Logout)
	authGroup.Post("/verify-email", authHandler.VerifyEmail)
	authGroup.Get("/verify-email", authHandler.VerifyEmailLink)
	authGroup.Post("/resend-verification", authHandler.ResendVerification)

	// Public merchandising reads (no auth) — the storefront browses signed-out.
	productHandler := handler.NewProductHandler(merchandising.Products, logger)
	skuHandler := handler.NewSKUHandler(merchandising.SKUs, logger)
	api.Get("/products", productHandler.Products)
	api.Get("/products/:id", productHandler.Product)
	api.Get("/skus", skuHandler.SKUs)
	api.Get("/brands", handler.NewBrandHandler(merchandising.Brands, logger).List)
	api.Get("/categories", handler.NewCategoryHandler(merchandising.Categories, logger).List)
	api.Get("/collections", handler.NewCollectionHandler(merchandising.Collections, logger).List)
	api.Get("/colourways", handler.NewColourwayHandler(merchandising.Colourways, logger).List)
	api.Get("/sizes", handler.NewSizeHandler(merchandising.Sizes, logger).List)

	protected := api.Group("", middleware.JWTAuth(tokens, users))
	protected.Get("/auth/me", authHandler.Me)

	userHandler := handler.NewUserHandler(users, auth, logger)
	protected.Patch("/users/me", userHandler.UpdateSelf)
	protected.Put("/users/me/password", userHandler.ChangePassword)
	protected.Delete("/users/me", userHandler.DeleteSelf)

	// Authenticated orders, scoped to the current user.
	orderHandler := handler.NewOrderHandler(orders, logger)
	protected.Post("/orders", orderHandler.Create)
	protected.Get("/orders", orderHandler.List)
	protected.Get("/orders/:id", orderHandler.Get)

	admin := protected.Group("/users", middleware.AdminOnly)
	admin.Get("/", userHandler.GetUsers)
	admin.Post("/", userHandler.CreateUser)
	admin.Get("/:id", userHandler.GetUserByID)
	admin.Put("/:id", userHandler.UpdateUser)
	admin.Put("/:id/role", userHandler.ChangeRole)
	admin.Delete("/:id", userHandler.DeleteUser)

	// Admin-only merchandising writes.
	adminCatalog := protected.Group("", middleware.AdminOnly)
	adminCatalog.Put("/admin/skus/:id/inventory", skuHandler.SetInventory)
	uploadHandler := handler.NewUploadHandler(uploader, logger)
	adminCatalog.Post("/admin/products/assets/batch", uploadHandler.ProductImages)
	adminCatalog.Post("/admin/products", productHandler.Create)
	adminCatalog.Put("/admin/products/:id", productHandler.Update)
	adminCatalog.Delete("/admin/products/:id", productHandler.Delete)
}
