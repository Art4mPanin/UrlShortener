package utils

import (
	"UrlShortener/internal/config"
	"fmt"
	"gopkg.in/gomail.v2"
	"log"
	"math/rand"
	"strconv"
	"time"
)

func VerificationCode(email string) (string, time.Time) {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}
	code, tstmp := CodeGeneration()
	m := gomail.NewMessage()
	m.SetHeader("From", cfg.Verification.EmailSender)
	m.SetHeader("To", email)
	m.SetHeader("Subject", "Verification Code")
	m.SetBody("text/plain", fmt.Sprintf("Thank you for registering!\nThis is a verification code: %v", code))
	d := gomail.NewDialer("smtp.gmail.com", 587, cfg.Verification.EmailServerUsername, cfg.Verification.EmailServerPassword)
	if err := d.DialAndSend(m); err != nil {
		log.Printf("Failed to send verification email: %v", err)
		return code, time.Time{} // Return an empty time object in case of error.
	}
	return code, tstmp
}

func CodeGeneration() (string, time.Time) {
	code := rand.Intn(1000000)
	timestamp := time.Now()
	return strconv.Itoa(code), timestamp
}

//meme
//Add timestamp
//add adding code to db
//
//
//
//add user activity check
//
//
//add otp validity check
//add otp correctness check
//
//
//
//add page after signing up
//add href
//add href- set user activity true
//
//После регистрации:
//setCookie("email", "...") - js
//редирект на вериф код
