package user

type defDBNamesCheckout_ struct {
	TableName string
	ID string
	PaymentIntentID string
	UserID string
	CreationTime string
	PaymentIntentStatusID string
	PaymentIntentStatus string
	HasNfts string
	
}

var DBNamesCheckout = &defDBNamesCheckout_{
	TableName: "checkouts",
	ID: "id",
	PaymentIntentID: "payment_intent_id",
	UserID: "user_id",
	CreationTime: "creation_time",
	PaymentIntentStatusID: "payment_intent_status_id",
	PaymentIntentStatus: "PaymentIntentStatus",
	HasNfts: "HasNfts",
}

type defDBNamesPaymentIntentStatus_ struct {
	TableName string
	ID string
	Name string
	
}

var DBNamesPaymentIntentStatus = &defDBNamesPaymentIntentStatus_{
	TableName: "payment_intent_statuses",
	ID: "id",
	Name: "name",
}

type defDBNamesCheckoutHasNfts_ struct {
	TableName string
	CheckoutID string
	NftID string
	Amount string
	
}

var DBNamesCheckoutHasNfts = &defDBNamesCheckoutHasNfts_{
	TableName: "checkout_has_nfts",
	CheckoutID: "checkout_id",
	NftID: "nft_id",
	Amount: "amount",
}

type defDBNamesDesignerApplication_ struct {
	TableName string
	UserID string
	SubmitTime string
	IP string
	User string
	
}

var DBNamesDesignerApplication = &defDBNamesDesignerApplication_{
	TableName: "designer_applications",
	UserID: "user_id",
	SubmitTime: "submit_time",
	IP: "ip",
	User: "User",
}

type defDBNamesProfile_ struct {
	TableName string
	UserID string
	Description string
	HasImage string
	
}

var DBNamesProfile = &defDBNamesProfile_{
	TableName: "profiles",
	UserID: "user_id",
	Description: "description",
	HasImage: "has_image",
}

type defDBNamesRole_ struct {
	TableName string
	ID string
	Name string
	Users string
	
}

var DBNamesRole = &defDBNamesRole_{
	TableName: "roles",
	ID: "id",
	Name: "name",
	Users: "users",
}

type defDBNamesStripeID_ struct {
	TableName string
	ID string
	UserID string
	StripeIDTypeID string
	StripeIDType string
	
}

var DBNamesStripeID = &defDBNamesStripeID_{
	TableName: "stripe_ids",
	ID: "id",
	UserID: "user_id",
	StripeIDTypeID: "stripe_id_type_id",
	StripeIDType: "StripeIDType",
}

type defDBNamesStripeIDType_ struct {
	TableName string
	ID string
	Name string
	
}

var DBNamesStripeIDType = &defDBNamesStripeIDType_{
	TableName: "stripe_id_types",
	ID: "id",
	Name: "name",
}

type defDBNamesUser_ struct {
	TableName string
	ID string
	FirstName string
	LastName string
	PreferredName string
	RegisterTime string
	Password string
	Email string
	Country string
	Roles string
	UserHasAddresses string
	Profile string
	UserHasNftsOffchain string
	StripeCustomer string
	HasCompleteProfile string
	HasBankAccount string
	HasUploadedOneNft string
	StripeTransferCapabilityStatus string
	
}

var DBNamesUser = &defDBNamesUser_{
	TableName: "users",
	ID: "id",
	FirstName: "first_name",
	LastName: "last_name",
	PreferredName: "preferred_name",
	RegisterTime: "register_time",
	Password: "password",
	Email: "email",
	Country: "country",
	Roles: "roles",
	UserHasAddresses: "UserHasAddresses",
	Profile: "Profile",
	UserHasNftsOffchain: "UserHasNftsOffchain",
	StripeCustomer: "StripeCustomer",
	HasCompleteProfile: "has_complete_profile",
	HasBankAccount: "has_bank_account",
	HasUploadedOneNft: "has_uploaded_one_nft",
	StripeTransferCapabilityStatus: "stripe_transfer_capability_status",
}

type defDBNamesUserHasAddresses_ struct {
	TableName string
	AddressID string
	UserID string
	
}

var DBNamesUserHasAddresses = &defDBNamesUserHasAddresses_{
	TableName: "user_has_addresses",
	AddressID: "address_id",
	UserID: "user_id",
}

type defDBNamesUserHasOffchainNfts_ struct {
	TableName string
	UserID string
	NftID string
	Amount string
	
}

var DBNamesUserHasOffchainNfts = &defDBNamesUserHasOffchainNfts_{
	TableName: "user_has_offchain_nfts",
	UserID: "user_id",
	NftID: "nft_id",
	Amount: "amount",
}

type defDBNamesUserHasRoles_ struct {
	TableName string
	UserID string
	RoleID string
	
}

var DBNamesUserHasRoles = &defDBNamesUserHasRoles_{
	TableName: "user_has_roles",
	UserID: "user_id",
	RoleID: "role_id",
}
