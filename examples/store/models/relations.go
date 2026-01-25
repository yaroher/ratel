package models

import (
	"github.com/yaroher/ratel/pkg/schema"
)

// ============================================================================
// One-to-Many Relations
// ============================================================================

// UsersOrders defines the one-to-many relationship between users and orders
// A user has many orders
var UsersOrders = schema.HasMany[UsersAlias, UsersColumnAlias, *UsersScanner, OrdersAlias, OrdersColumnAlias, *OrdersScanner](
	UsersAliasName,
	OrdersRef,
	OrdersColumnUserID,
	UsersColumnUserID,
)

// OrdersUser defines the belongs-to relationship between orders and users
// An order belongs to a user
var OrdersUser = schema.BelongsTo[OrdersAlias, OrdersColumnAlias, *OrdersScanner, UsersAlias, UsersColumnAlias, *UsersScanner](
	OrdersAliasName,
	UsersRef,
	OrdersColumnUserID,
	UsersColumnUserID,
)

// OrdersCurrency defines the belongs-to relationship between orders and currency
// An order belongs to a currency
var OrdersCurrency = schema.BelongsTo[OrdersAlias, OrdersColumnAlias, *OrdersScanner, CurrencyAlias, CurrencyColumnAlias, *CurrencyScanner](
	OrdersAliasName,
	CurrencyRef,
	OrdersColumnCurrency,
	CurrencyColumnCode,
)

// ============================================================================
// Many-to-Many Relations
// ============================================================================

// ProductsCategories defines the many-to-many relationship between products and categories
// A product can belong to many categories, and a category can have many products
var ProductsCategories = schema.ManyToMany[
	ProductsAlias, ProductsColumnAlias, *ProductsScanner,
	ProductCategoriesAlias, ProductCategoriesColumnAlias,
	CategoriesAlias, CategoriesColumnAlias, *CategoriesScanner,
](
	ProductsAliasName,
	ProductCategoriesRef,
	CategoriesRef,
	ProductsColumnProductID,           // local key in products
	ProductCategoriesColumnProductID,  // pivot foreign key to products
	ProductCategoriesColumnCategoryID, // pivot foreign key to categories
	CategoriesColumnCategoryID,        // related key in categories
)

// CategoriesProducts defines the inverse many-to-many relationship
// A category has many products through product_categories
var CategoriesProducts = schema.ManyToMany[
	CategoriesAlias, CategoriesColumnAlias, *CategoriesScanner,
	ProductCategoriesAlias, ProductCategoriesColumnAlias,
	ProductsAlias, ProductsColumnAlias, *ProductsScanner,
](
	CategoriesAliasName,
	ProductCategoriesRef,
	ProductsRef,
	CategoriesColumnCategoryID,        // local key in categories
	ProductCategoriesColumnCategoryID, // pivot foreign key to categories
	ProductCategoriesColumnProductID,  // pivot foreign key to products
	ProductsColumnProductID,           // related key in products
)

// ProductsTags defines the many-to-many relationship between products and tags
// A product can have many tags, and a tag can be on many products
var ProductsTags = schema.ManyToMany[
	ProductsAlias, ProductsColumnAlias, *ProductsScanner,
	ProductTagsAlias, ProductTagsColumnAlias,
	TagsAlias, TagsColumnAlias, *TagsScanner,
](
	ProductsAliasName,
	ProductTagsRef,
	TagsRef,
	ProductsColumnProductID,    // local key in products
	ProductTagsColumnProductID, // pivot foreign key to products
	ProductTagsColumnTagID,     // pivot foreign key to tags
	TagsColumnTagID,            // related key in tags
)

// TagsProducts defines the inverse many-to-many relationship
// A tag has many products through product_tags
var TagsProducts = schema.ManyToMany[
	TagsAlias, TagsColumnAlias, *TagsScanner,
	ProductTagsAlias, ProductTagsColumnAlias,
	ProductsAlias, ProductsColumnAlias, *ProductsScanner,
](
	TagsAliasName,
	ProductTagsRef,
	ProductsRef,
	TagsColumnTagID,            // local key in tags
	ProductTagsColumnTagID,     // pivot foreign key to tags
	ProductTagsColumnProductID, // pivot foreign key to products
	ProductsColumnProductID,    // related key in products
)

// ============================================================================
// Compile-time interface checks
// ============================================================================

var _ schema.RelationTableAlias[CurrencyAlias] = Currency.Table
var _ schema.RelationTableAlias[UsersAlias] = Users.Table
var _ schema.RelationTableAlias[OrdersAlias] = Orders.Table
var _ schema.RelationTableAlias[ProductsAlias] = Products.Table
var _ schema.RelationTableAlias[OrderItemsAlias] = OrderItems.Table
var _ schema.RelationTableAlias[CategoriesAlias] = Categories.Table
var _ schema.RelationTableAlias[TagsAlias] = Tags.Table
var _ schema.RelationTableAlias[ProductCategoriesAlias] = ProductCategories.Table
var _ schema.RelationTableAlias[ProductTagsAlias] = ProductTags.Table

var _ schema.RelationTableJoin[OrdersAlias, OrdersColumnAlias] = Orders.Table
var _ schema.RelationTableQuery[OrdersAlias, OrdersColumnAlias, *OrdersScanner] = Orders.Table
var _ schema.RelationTableQuery[CategoriesAlias, CategoriesColumnAlias, *CategoriesScanner] = Categories.Table
var _ schema.RelationTableQuery[TagsAlias, TagsColumnAlias, *TagsScanner] = Tags.Table
var _ schema.RelationTableQuery[ProductsAlias, ProductsColumnAlias, *ProductsScanner] = Products.Table
