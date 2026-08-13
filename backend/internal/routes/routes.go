package routes

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/yven/backend/internal/config"
	"github.com/yven/backend/internal/handlers"
	"github.com/yven/backend/internal/middleware"
	"github.com/yven/backend/internal/models"
	"gorm.io/gorm"
)

func Register(r *gin.Engine, cfg config.Config, db *gorm.DB) {
	r.GET("/healthz", func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"status": "ok"}) })

	volunteerH := &handlers.VolunteerHandler{DB: db}
	serviceLogH := &handlers.ServiceLogHandler{DB: db}
	verificationH := &handlers.VerificationHandler{DB: db}
	vsrH := &handlers.VSRHandler{DB: db}
	eventH := &handlers.EventHandler{DB: db}

	api := r.Group("/api/v1")
	{
		// Public — no auth required.
		api.POST("/volunteers/register", volunteerH.Register)
		api.GET("/events", eventH.List)
		api.GET("/events/:id", eventH.Get)

		// Everything below requires a valid Auth0 session.
		authed := api.Group("")
		authed.Use(middleware.RequireAuth(cfg, db))
		{
			// Volunteer-only routes.
			vol := authed.Group("")
			vol.Use(middleware.RequireRole(models.RoleVolunteer))
			{
				vol.GET("/me/dashboard", volunteerH.Dashboard)
				vol.PATCH("/me/profile", volunteerH.UpdateProfile)

				vol.POST("/service-logs", serviceLogH.CreateDraft)
				vol.POST("/service-logs/:id/submit", serviceLogH.Submit)
				vol.DELETE("/service-logs/:id", serviceLogH.Withdraw)

				vol.GET("/me/vsr", vsrH.GetMyVSR)
				vol.GET("/me/vsr/export", vsrH.Export)

				vol.POST("/events/:id/register", eventH.Register)
			}

			// Supervisor-only routes (event supervisors + org admins can
			// both hold verification authority; the handler itself does
			// the fine-grained org-membership check).
			sup := authed.Group("")
			sup.Use(middleware.RequireRole(models.RoleOrgSupervisor, models.RoleOrgAdmin))
			{
				sup.GET("/verification/queue", verificationH.Queue)
				sup.POST("/service-logs/:id/decide", verificationH.Decide)
				sup.POST("/organizations/:orgId/events", eventH.Create)
				sup.POST("/events/:id/publish", eventH.Publish)
			}

			// YVEN Admin routes — see docs/ROADMAP.md for the remaining
			// Conflict Dashboard / Policy Management / oversight-dashboard
			// endpoints still to be built out on top of the models already
			// in place (models/admin.go).
			admin := authed.Group("/admin")
			admin.Use(middleware.RequireRole(models.RoleYVENAdmin))
			{
				admin.GET("/healthcheck", func(c *gin.Context) {
					c.JSON(http.StatusOK, gin.H{"status": "admin routes online — see docs/ROADMAP.md"})
				})
			}
		}
	}
}
