package entity

import (
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"game/api/internal/infra/database"
)

type Wallet struct {
	guid    uuid.UUID
	balance float64
}

type Client struct {
	id       uuid.UUID
	username string
	password string
	inPlay   bool
	wallet   Wallet
}

func NewClient(username, password string) (Client, error) {
	hashedPassword, err := HashPassword(password)
	if err != nil {
		return Client{}, err
	}

	return Client{
		id:       uuid.New(),
		username: username,
		password: hashedPassword,
		inPlay:   false,
		wallet: Wallet{
			guid:    uuid.New(),
			balance: 0,
		},
	}, nil
}

func (c *Client) GetID() uuid.UUID {
	return c.id
}

func (c *Client) GetUsername() string {
	return c.username
}

func (c *Client) GetPassword() string {
	return c.password
}

func (c *Client) GetBalance() float64 {
	return c.wallet.balance
}

func (c *Client) CanBet(amount float64) bool {
	return c.wallet.balance >= amount
}

func (c *Client) InPlay() bool {
	return c.inPlay
}

func (c *Client) Debit(amount float64) {
	c.wallet.balance -= amount
}

func (c *Client) Credit(amount float64) {
	c.wallet.balance += amount
}

func (c *Client) PlayOn() {
	c.inPlay = true
}

func (c *Client) PlayOff() {
	c.inPlay = true
}

func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

func LoadClient(cData database.ClientData) (c Client, err error) {
	c.id, err = uuid.Parse(cData.GUID)
	if err != nil {
		return
	}
	c.username = cData.Username
	c.password = cData.Password
	wGuid, err := uuid.Parse(cData.Wallet.GUID)
	if err != nil {
		return
	}
	c.wallet = Wallet{
		guid:    wGuid,
		balance: cData.Wallet.Balance,
	}
	return
}
