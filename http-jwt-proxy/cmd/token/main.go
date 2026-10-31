// Command token creates short-lived development tokens for testing the proxy.
package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)
func main() {
	log.SetFlags(0)
	subject:=flag.String("subject","student","value of the JWT sub claim")
	ttl:=flag.Duration("ttl",15*time.Minute,"token lifetime, for example 15m or -1m")
	flag.Parse()
	secret:=os.Getenv("JWT_SECRET")
	if len(secret)<32 { log.Fatal("JWT_SECRET must contain at least 32 bytes") }
	issuer:=envOrDefault("JWT_ISSUER","http-hands-on")
	audience:=envOrDefault("JWT_AUDIENCE","legacy-api")
	now:=time.Now().UTC()
	claims:=jwt.RegisteredClaims{Issuer:issuer,Subject:*subject,Audience:jwt.ClaimStrings{audience},ExpiresAt:jwt.NewNumericDate(now.Add(*ttl)),NotBefore:jwt.NewNumericDate(now.Add(-5*time.Second)),IssuedAt:jwt.NewNumericDate(now)}
	token:=jwt.NewWithClaims(jwt.SigningMethodHS256,claims)
	signed,err:=token.SignedString([]byte(secret)); if err!=nil { log.Fatalf("sign token: %v",err) }
	fmt.Println(signed)
}
func envOrDefault(key,fallback string)string { if value:=os.Getenv(key); value!="" { return value }; return fallback }
