package middleware

import (
	"fmt"
	"strings"

	"continental-backend/internal/config"
	"continental-backend/internal/utils"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

type JWTClaims struct {
	ID     int    `json:"id"`
	Rol    string `json:"rol"`
	Nombre string `json:"nombre"`
	jwt.RegisteredClaims
}

func VerifyToken() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			utils.RespondError(c, utils.NewUnauthorizedError("Token no proporcionado"))
			c.Abort()
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")

		token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("método de firma no esperado: %v", token.Header["alg"])
			}
			return []byte(config.AppConfig.JWTSecret), nil
		})

		if err != nil || !token.Valid {
			utils.RespondError(c, utils.NewUnauthorizedError("Token inválido o expirado"))
			c.Abort()
			return
		}

		claims, ok := token.Claims.(*JWTClaims)
		if !ok {
			utils.RespondError(c, utils.NewUnauthorizedError("Token inválido"))
			c.Abort()
			return
		}

		c.Set("userId", claims.ID)
		c.Set("userRol", claims.Rol)
		c.Set("userNombre", claims.Nombre)

		c.Next()
	}
}

func RequireRole(rolesPermitidos ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userRol, exists := c.Get("userRol")
		if !exists {
			utils.RespondError(c, utils.NewUnauthorizedError("Debe autenticarse primero"))
			c.Abort()
			return
		}

		rolStr, ok := userRol.(string)
		if !ok {
			utils.RespondError(c, utils.NewUnauthorizedError("Rol inválido"))
			c.Abort()
			return
		}

		permitido := false
		for _, rol := range rolesPermitidos {
			if rol == rolStr {
				permitido = true
				break
			}
		}

		if !permitido {
			utils.RespondError(c, utils.NewForbiddenError(fmt.Sprintf("Acceso restringido. Roles permitidos: %s", strings.Join(rolesPermitidos, ", "))))
			c.Abort()
			return
		}

		c.Next()
	}
}

// Helpers para extraer usuario del contexto
func GetUserID(c *gin.Context) int {
	if val, exists := c.Get("userId"); exists {
		if id, ok := val.(int); ok {
			return id
		}
	}
	return 0
}

func GetUserRol(c *gin.Context) string {
	if val, exists := c.Get("userRol"); exists {
		if rol, ok := val.(string); ok {
			return rol
		}
	}
	return ""
}

func GetUserNombre(c *gin.Context) string {
	if val, exists := c.Get("userNombre"); exists {
		if nombre, ok := val.(string); ok {
			return nombre
		}
	}
	return ""
}
