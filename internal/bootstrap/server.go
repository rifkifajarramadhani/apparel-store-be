package bootstrap

import (
	"log/slog"
	"time"

	"github.com/rifkifajarramadhani/golang-clean-architecture/internal/adapter/jobs"
	"github.com/rifkifajarramadhani/golang-clean-architecture/internal/adapter/jwt"
	mysqladapter "github.com/rifkifajarramadhani/golang-clean-architecture/internal/adapter/mysql"
	"github.com/rifkifajarramadhani/golang-clean-architecture/internal/adapter/password"
	queueadapter "github.com/rifkifajarramadhani/golang-clean-architecture/internal/adapter/queue"
	"github.com/rifkifajarramadhani/golang-clean-architecture/internal/auth"
	"github.com/rifkifajarramadhani/golang-clean-architecture/internal/brand"
	"github.com/rifkifajarramadhani/golang-clean-architecture/internal/category"
	"github.com/rifkifajarramadhani/golang-clean-architecture/internal/collection"
	"github.com/rifkifajarramadhani/golang-clean-architecture/internal/colourway"
	"github.com/rifkifajarramadhani/golang-clean-architecture/internal/config"
	appmail "github.com/rifkifajarramadhani/golang-clean-architecture/internal/mail"
	"github.com/rifkifajarramadhani/golang-clean-architecture/internal/order"
	"github.com/rifkifajarramadhani/golang-clean-architecture/internal/product"
	"github.com/rifkifajarramadhani/golang-clean-architecture/internal/queue"
	appsize "github.com/rifkifajarramadhani/golang-clean-architecture/internal/size"
	"github.com/rifkifajarramadhani/golang-clean-architecture/internal/sku"
	"github.com/rifkifajarramadhani/golang-clean-architecture/internal/user"
	"gorm.io/gorm"
)

type HTTPServices struct {
	Users       *user.Service
	Auth        *auth.Service
	Products    *product.Service
	SKUs        *sku.Service
	Brands      *brand.Service
	Categories  *category.Service
	Collections *collection.Service
	Colourways  *colourway.Service
	Sizes       *appsize.Service
	Orders      *order.Service
	Tokens      *jwt.Service
}

func WireHTTPServices(cfg *config.Config, db *gorm.DB, logger *slog.Logger, dispatcher queue.Dispatcher) HTTPServices {
	repository := mysqlRepository(db)
	hasher := password.Bcrypt{}
	users := user.NewService(repository, hasher)
	tokens := jwt.NewService(
		cfg.Auth.JWTAccessSecret,
		cfg.Auth.JWTRefreshSecret,
		cfg.Auth.AccessTTLMinutes,
		cfg.Auth.RefreshTTLHours,
		cfg.Auth.Issuer,
		cfg.Auth.Audience,
	)
	mailer := appmail.NewMailer(
		appmail.Address{Name: cfg.Mail.FromName, Address: cfg.Mail.FromAddress},
		nil,
		queueadapter.NewMailDispatcher(dispatcher),
	)
	authService := auth.NewService(
		repository,
		repository,
		repository,
		tokens,
		hasher,
		jobs.NewVerificationNotifier(mailer, logger, cfg.App.PublicURL),
		jobs.NewWelcomeNotifier(mailer, logger, cfg.App.StorefrontURL),
		time.Duration(cfg.Auth.VerificationTTLHours)*time.Hour,
		cfg.Auth.BootstrapAdminEmail,
	)
	products := product.NewService(mysqladapter.NewProductRepository(db))
	skus := sku.NewService(mysqladapter.NewSKURepository(db))
	orderService := order.NewService(mysqladapter.NewOrderRepository(db))
	return HTTPServices{
		Users: users, Auth: authService, Products: products, SKUs: skus,
		Brands:      brand.NewService(mysqladapter.NewBrandRepository(db)),
		Categories:  category.NewService(mysqladapter.NewCategoryRepository(db)),
		Collections: collection.NewService(mysqladapter.NewCollectionRepository(db)),
		Colourways:  colourway.NewService(mysqladapter.NewColourwayRepository(db)),
		Sizes:       appsize.NewService(mysqladapter.NewSizeRepository(db)),
		Orders:      orderService, Tokens: tokens,
	}
}
