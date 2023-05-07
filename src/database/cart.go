package database

import "errors"

var (
	ErrCantFindProduct    = errors.New("can't find the product")
	ErrCantDecodeProducts = errors.New("can't decode product")
	ErrUserIdIsNotVAlid   = errors.New("user is not valid")
	ErrCantUpdateUser     = errors.New("can't add this product to the cart")
	ErrCantRemoveItemCart = errors.New("can't remove this item from cart")
	ErrCantGetItem        = errors.New("can't get item")
	ErrCantBuyCartItem    = errors.New("can't buy cart item")
)

func AddProductToCart() {

}

func RemoveCartItem() {

}

func BuyItemFromCart() {

}

func IstantBuyer() {

}
