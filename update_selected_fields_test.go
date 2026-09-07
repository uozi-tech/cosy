package cosy

import (
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/uozi-tech/cosy/model"
	"github.com/uozi-tech/cosy/router"
	"github.com/uozi-tech/cosy/sandbox"
)

// TestModifySelectsOnlySubmittedFields pins the update contract: only the
// columns present in the request body are written, a submitted zero value is
// written, keys without an update rule are ignored, and a body with no
// updatable field is rejected with 406 instead of wiping the row.
func TestModifySelectsOnlySubmittedFields(t *testing.T) {
	sandbox.NewInstance("app.ini", "pgsql").
		RegisterModels(User{}).
		Run(func(instance *sandbox.Instance) {
			r := router.GetEngine()
			g := r.Group("/")
			api := Api[User]("users")

			var hookSelected []string
			api.ModifyHook(func(c *Ctx[User]) {
				// Declared before the payload exists: must only take effect
				// when the client actually submits the field.
				c.AddSelectedFields("phone")
				c.BeforeExecuteHook(func(ctx *Ctx[User]) {
					hookSelected = ctx.GetSelectedFields()
				})
			})
			api.InitRouter(g)

			client := instance.GetClient()
			userID := testCreate(t, instance)
			require.NotEmpty(t, userID)

			db := model.UseDB(instance.Context())
			load := func() User {
				var u User
				require.NoError(t, db.First(&u, "id = ?", userID).Error)
				return u
			}
			before := load()

			// 1. partial body: name changes, bio is cleared, the rest is untouched
			resp, err := client.Post("/users/"+userID, gin.H{
				"name":        "partial-name",
				"bio":         "",
				"last_active": "2024-03-13T11:22:44Z", // no update rule -> dropped
			})
			require.NoError(t, err)
			require.Equal(t, 200, resp.StatusCode, resp.BodyText())

			after := load()
			assert.Equal(t, "partial-name", after.Name)
			assert.Equal(t, "", after.Bio)
			assert.Equal(t, before.Email, after.Email)
			assert.Equal(t, before.Phone, after.Phone)
			assert.Equal(t, before.Age, after.Age)
			assert.Equal(t, before.Title, after.Title)
			assert.Equal(t, before.SchoolID, after.SchoolID)
			assert.Equal(t, before.Password, after.Password)
			assert.Equal(t, before.Status, after.Status)
			assert.Nil(t, after.LastActive)
			assert.ElementsMatch(t, []string{"name", "bio"}, hookSelected,
				"phone was declared as a candidate but not submitted")

			// 2. candidate field submitted -> selected once, no duplicates
			resp, err = client.Post("/users/"+userID, gin.H{"phone": "13800000000"})
			require.NoError(t, err)
			require.Equal(t, 200, resp.StatusCode, resp.BodyText())
			assert.Equal(t, "13800000000", load().Phone)
			assert.ElementsMatch(t, []string{"phone"}, hookSelected)

			// 3. a body with no updatable field is rejected and the row is untouched
			for _, body := range []gin.H{
				{},
				{"last_active": "2024-03-13T11:22:44Z"}, // present but without an update rule
			} {
				resp, err = client.Post("/users/"+userID, body)
				require.NoError(t, err)
				assert.Equal(t, 406, resp.StatusCode, resp.BodyText())
				var verr ValidateError
				require.NoError(t, resp.To(&verr))
				assert.Equal(t, ErrEmptyPayload, verr.Errors["body"])
			}

			after = load()
			assert.Equal(t, "partial-name", after.Name)
			assert.Equal(t, before.Email, after.Email)
			assert.Equal(t, before.Password, after.Password)
			assert.Equal(t, "13800000000", after.Phone)
		})
}
