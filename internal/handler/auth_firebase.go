package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	models "my_app/internal/model"
	utils "my_app/internal/util"
	"net/http"
	"time"

	"firebase.google.com/go/v4/auth"
	"github.com/golang-jwt/jwt/v4"
	_ "github.com/lib/pq" // Import PostgreSQL driver
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var (
	jwtSecret = []byte("serect_key")
)

func Login(w http.ResponseWriter, r *http.Request, authClient *auth.Client, ctx context.Context) {
	var req struct {
		Token string `json:"token"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}
	if req.Token == "" {
		http.Error(w, "Missing token parameter", http.StatusBadRequest)
		return
	}

	verifyTokenResp, err := authClient.VerifyIDToken(ctx, req.Token)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error verifying token: %v", err), http.StatusInternalServerError)
		return
	}

	email_id := verifyTokenResp.UID
	emailUser, ok := verifyTokenResp.Claims["email"].(string)
	if !ok {
		http.Error(w, "Error extracting email from token", http.StatusInternalServerError)
		return
	}

	user := bson.M{
		"email_id":      email_id,
		"email":         emailUser,
		"password_hash": utils.HashPassword("DefaultPassword"),
		"first_name":    "Default",
		"last_name":     "Default",
		"created_at":    time.Now(),
	}

	repoUser := models.NewQuestionRepository("users")
	user_row, err := repoUser.GetFilter(bson.M{"email_id": email_id})
	var user_id primitive.ObjectID
	if user_row == nil {
		user_id, err = repoUser.Create(user)
		if err != nil {
			return
		}
	} else {
		user_id = user_row["_id"].(primitive.ObjectID)
	}

	token, err := CreateJWT(user_id, email_id, emailUser)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to create JWT: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"token": token})
}

func CreateJWT(user_id primitive.ObjectID, acc_id, email string) (string, error) {
	claims := jwt.MapClaims{
		"_id":      user_id,
		"email_id": acc_id,
		"email":    email,
		"exp":      time.Now().Add(time.Hour * 24).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	jwtToken, err := token.SignedString(jwtSecret)
	if err != nil {
		return "", fmt.Errorf("failed to sign JWT token: %v", err)
	}
	return jwtToken, nil
}

func ValidateJWT(tokenString string) (*jwt.Token, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return jwtSecret, nil
	})

	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	return token, nil
}

func CheckTokenExpiry(token *jwt.Token) error {
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return fmt.Errorf("failed to get claims")
	}

	exp, ok := claims["exp"].(float64)
	if !ok {
		return fmt.Errorf("failed to get exp claim")
	}

	if time.Now().Unix() > int64(exp) {
		return fmt.Errorf("token has expired")
	}

	return nil
}

func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tokenString := r.Header.Get("Authorization")
		if tokenString == "" {
			http.Error(w, "Missing token", http.StatusUnauthorized)
			return
		}

		// Remove "Bearer " prefix if present
		if len(tokenString) > 7 && tokenString[:7] == "Bearer " {
			tokenString = tokenString[7:]
		}

		token, err := ValidateJWT(tokenString)
		if err != nil {
			http.Error(w, fmt.Sprintf("Invalid token: %v", err), http.StatusUnauthorized)
			return
		}

		err = CheckTokenExpiry(token)
		if err != nil {
			http.Error(w, fmt.Sprintf("Token expired: %v", err), http.StatusUnauthorized)
			return
		}

		claims, _ := token.Claims.(jwt.MapClaims)
		fmt.Println(claims["_id"])
		ctx := context.WithValue(r.Context(), "_id", claims["_id"])
		ctx = context.WithValue(ctx, "email_id", claims["email_id"])
		ctx = context.WithValue(ctx, "email", claims["email"])

		next.ServeHTTP(w, r.WithContext(ctx))

	})
}
