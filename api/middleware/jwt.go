package middleware

import (
	"errors"

	"github.com/bitrainforest/filmeta-hic/core/assert"

	"github.com/go-kratos/kratos/v2/config"

	"net/http"
	"time"

	"github.com/bitrainforest/pulsar/internal/service"

	"github.com/bitrainforest/pulsar/api/req"

	jwt "github.com/appleboy/gin-jwt/v2"
	"github.com/gin-gonic/gin"
)

const (
	TimeFormat = "2006-01-02 15:04:05"
)

var (
	secret                string
	identityKey           = "appId"
	ErrMissApplyInfoErr   = errors.New("the appId or appSecret parameter is missing")
	ErrFailedApplyInfoErr = errors.New("appId or appSecret illegal parameters")
	ErrFailedTokenErr     = errors.New("token is illegal or invalid, please check your token")
)

type (
	ApplyCx struct {
		AppId string `json:"appId"`
	}
	JWTMiddleware struct {
		appService service.UserAppService
	}
)

func MustLoadSecret(conf config.Config) {
	var (
		err error
	)
	v := conf.Value("data.jwt.secret")
	secret, err = v.String()
	if err != nil {
		assert.CheckErr(err)
	}
}

func NewJWTMiddleware(appService service.UserAppService) *JWTMiddleware {
	return &JWTMiddleware{appService: appService}
}

func NewAuthMiddleware(middleware *JWTMiddleware) (*jwt.GinJWTMiddleware, error) {
	return jwt.New(&jwt.GinJWTMiddleware{
		Realm:           "nuwas",
		Key:             []byte(secret),
		Timeout:         time.Hour * 2,
		MaxRefresh:      time.Hour,
		IdentityKey:     identityKey,
		PayloadFunc:     PayloadFunc(),
		LoginResponse:   LoginResponse(),
		IdentityHandler: IdentityHandler(),
		Authenticator:   Authenticator(middleware),
		Unauthorized:    Unauthorized(),
		RefreshResponse: RefreshResponse(),
		TokenLookup:     "header: Authorization, query: token, cookie: jwt",
		TokenHeadName:   "Bearer",
		TimeFunc:        time.Now,
	})
}

func PayloadFunc() func(data interface{}) jwt.MapClaims {
	return func(data interface{}) jwt.MapClaims {
		if v, ok := data.(*ApplyCx); ok {
			return jwt.MapClaims{
				identityKey: v.AppId,
			}
		}
		return jwt.MapClaims{}
	}
}

func LoginResponse() func(*gin.Context, int, string, time.Time) {
	return func(c *gin.Context, i int, token string, t time.Time) {
		c.JSON(http.StatusOK, gin.H{
			"code":   http.StatusOK,
			"token":  token,
			"expire": t.Format(TimeFormat),
		})
	}
}

func IdentityHandler() func(c *gin.Context) interface{} {
	return func(c *gin.Context) interface{} {
		claims := jwt.ExtractClaims(c)
		return &ApplyCx{
			AppId: claims[identityKey].(string),
		}
	}
}

func Authenticator(middleware *JWTMiddleware) func(c *gin.Context) (interface{}, error) {
	return func(c *gin.Context) (interface{}, error) {
		var param req.AppReq
		if err := c.ShouldBindJSON(&param); err != nil {
			return nil, ErrMissApplyInfoErr
		}

		err := func(server service.UserAppService) error {
			applyInfo, err := server.GetAppByAppId(c, param.AppId)
			if err != nil {
				return err
			}
			if applyInfo.IsEmpty() || !applyInfo.CheckAppSecret(param.AppSecret) {
				return ErrFailedApplyInfoErr
			}
			return nil
		}(middleware.appService)
		if err != nil {
			return nil, err
		}
		return &ApplyCx{AppId: param.AppId}, nil
	}
}

func Unauthorized() func(c *gin.Context, code int, message string) {
	return func(c *gin.Context, code int, message string) {
		c.JSON(http.StatusOK, gin.H{
			"code":    code,
			"message": message,
		})
	}
}

func RefreshResponse() func(*gin.Context, int, string, time.Time) {
	return func(c *gin.Context, i int, token string, t time.Time) {
		c.JSON(http.StatusOK, gin.H{
			"code":   http.StatusOK,
			"token":  token,
			"expire": t.Format(TimeFormat),
		})
	}
}
