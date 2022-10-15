package repo

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/hay-kot/homebox/backend/ent"
	"github.com/hay-kot/homebox/backend/ent/group"
	"github.com/hay-kot/homebox/backend/ent/item"
	"github.com/hay-kot/homebox/backend/ent/label"
	"github.com/hay-kot/homebox/backend/ent/location"
	"github.com/hay-kot/homebox/backend/ent/predicate"
)

type ItemsRepository struct {
	db *ent.Client
}

type (
	ItemQuery struct {
		Page        int
		PageSize    int
		Search      string      `json:"search"`
		LocationIDs []uuid.UUID `json:"locationIds"`
		LabelIDs    []uuid.UUID `json:"labelIds"`
		SortBy      string      `json:"sortBy"`
	}

	ItemCreate struct {
		ImportRef   string `json:"-"`
		Name        string `json:"name"`
		Description string `json:"description"`

		// Edges
		LocationID uuid.UUID   `json:"locationId"`
		LabelIDs   []uuid.UUID `json:"labelIds"`
	}
	ItemUpdate struct {
		ID          uuid.UUID `json:"id"`
		Name        string    `json:"name"`
		Description string    `json:"description"`
		Quantity    int       `json:"quantity"`
		Insured     bool      `json:"insured"`

		// Edges
		LocationID uuid.UUID   `json:"locationId"`
		LabelIDs   []uuid.UUID `json:"labelIds"`

		// Identifications
		SerialNumber string `json:"serialNumber"`
		ModelNumber  string `json:"modelNumber"`
		Manufacturer string `json:"manufacturer"`

		// Warranty
		LifetimeWarranty bool      `json:"lifetimeWarranty"`
		WarrantyExpires  time.Time `json:"warrantyExpires"`
		WarrantyDetails  string    `json:"warrantyDetails"`

		// Purchase
		PurchaseTime  time.Time `json:"purchaseTime"`
		PurchaseFrom  string    `json:"purchaseFrom"`
		PurchasePrice float64   `json:"purchasePrice,string"`

		// Sold
		SoldTime  time.Time `json:"soldTime"`
		SoldTo    string    `json:"soldTo"`
		SoldPrice float64   `json:"soldPrice,string"`
		SoldNotes string    `json:"soldNotes"`

		// Extras
		Notes string `json:"notes"`
		// Fields []*FieldSummary `json:"fields"`
	}

	ItemSummary struct {
		ImportRef   string    `json:"-"`
		ID          uuid.UUID `json:"id"`
		Name        string    `json:"name"`
		Description string    `json:"description"`
		Quantity    int       `json:"quantity"`
		Insured     bool      `json:"insured"`
		CreatedAt   time.Time `json:"createdAt"`
		UpdatedAt   time.Time `json:"updatedAt"`

		// Edges
		Location LocationSummary `json:"location"`
		Labels   []LabelSummary  `json:"labels"`
	}

	ItemOut struct {
		ItemSummary

		SerialNumber string `json:"serialNumber"`
		ModelNumber  string `json:"modelNumber"`
		Manufacturer string `json:"manufacturer"`

		// Warranty
		LifetimeWarranty bool      `json:"lifetimeWarranty"`
		WarrantyExpires  time.Time `json:"warrantyExpires"`
		WarrantyDetails  string    `json:"warrantyDetails"`

		// Purchase
		PurchaseTime  time.Time `json:"purchaseTime"`
		PurchaseFrom  string    `json:"purchaseFrom"`
		PurchasePrice float64   `json:"purchasePrice,string"`

		// Sold
		SoldTime  time.Time `json:"soldTime"`
		SoldTo    string    `json:"soldTo"`
		SoldPrice float64   `json:"soldPrice,string"`
		SoldNotes string    `json:"soldNotes"`

		// Extras
		Notes string `json:"notes"`

		Attachments []ItemAttachment `json:"attachments"`
		// Future
		// Fields []*FieldSummary `json:"fields"`
	}
)

var (
	mapItemsSummaryErr = mapTEachErrFunc(mapItemSummary)
)

func mapItemSummary(item *ent.Item) ItemSummary {
	var location LocationSummary
	if item.Edges.Location != nil {
		location = mapLocationSummary(item.Edges.Location)
	}

	var labels []LabelSummary
	if item.Edges.Label != nil {
		labels = mapEach(item.Edges.Label, mapLabelSummary)
	}

	return ItemSummary{
		ID:          item.ID,
		Name:        item.Name,
		Description: item.Description,
		Quantity:    item.Quantity,
		CreatedAt:   item.CreatedAt,
		UpdatedAt:   item.UpdatedAt,

		// Edges
		Location: location,
		Labels:   labels,

		// Warranty
		Insured: item.Insured,
	}
}

var (
	mapItemOutErr = mapTErrFunc(mapItemOut)
)

func mapItemOut(item *ent.Item) ItemOut {
	var attachments []ItemAttachment
	if item.Edges.Attachments != nil {
		attachments = mapEach(item.Edges.Attachments, ToItemAttachment)
	}

	return ItemOut{
		ItemSummary:      mapItemSummary(item),
		LifetimeWarranty: item.LifetimeWarranty,
		WarrantyExpires:  item.WarrantyExpires,
		WarrantyDetails:  item.WarrantyDetails,

		// Identification
		SerialNumber: item.SerialNumber,
		ModelNumber:  item.ModelNumber,
		Manufacturer: item.Manufacturer,

		// Purchase
		PurchaseTime:  item.PurchaseTime,
		PurchaseFrom:  item.PurchaseFrom,
		PurchasePrice: item.PurchasePrice,

		// Sold
		SoldTime:  item.SoldTime,
		SoldTo:    item.SoldTo,
		SoldPrice: item.SoldPrice,
		SoldNotes: item.SoldNotes,

		// Extras
		Notes:       item.Notes,
		Attachments: attachments,
	}
}

func (e *ItemsRepository) getOne(ctx context.Context, where ...predicate.Item) (ItemOut, error) {
	q := e.db.Item.Query().Where(where...)

	return mapItemOutErr(q.
		WithFields().
		WithLabel().
		WithLocation().
		WithGroup().
		WithAttachments(func(aq *ent.AttachmentQuery) {
			aq.WithDocument()
		}).
		Only(ctx),
	)
}

// GetOne returns a single item by ID. If the item does not exist, an error is returned.
// See also: GetOneByGroup to ensure that the item belongs to a specific group.
func (e *ItemsRepository) GetOne(ctx context.Context, id uuid.UUID) (ItemOut, error) {
	return e.getOne(ctx, item.ID(id))
}

// GetOneByGroup returns a single item by ID. If the item does not exist, an error is returned.
// GetOneByGroup ensures that the item belongs to a specific group.
func (e *ItemsRepository) GetOneByGroup(ctx context.Context, gid, id uuid.UUID) (ItemOut, error) {
	return e.getOne(ctx, item.ID(id), item.HasGroupWith(group.ID(gid)))
}

// QueryByGroup returns a list of items that belong to a specific group based on the provided query.
func (e *ItemsRepository) QueryByGroup(ctx context.Context, gid uuid.UUID, q ItemQuery) (PaginationResult[ItemSummary], error) {
	qb := e.db.Item.Query().Where(item.HasGroupWith(group.ID(gid)))

	if len(q.LabelIDs) > 0 {
		labels := make([]predicate.Item, 0, len(q.LabelIDs))
		for _, l := range q.LabelIDs {
			labels = append(labels, item.HasLabelWith(label.ID(l)))
		}
		qb = qb.Where(item.Or(labels...))
	}

	if len(q.LocationIDs) > 0 {
		locations := make([]predicate.Item, 0, len(q.LocationIDs))
		for _, l := range q.LocationIDs {
			locations = append(locations, item.HasLocationWith(location.ID(l)))
		}
		qb = qb.Where(item.Or(locations...))
	}

	if q.Search != "" {
		qb.Where(
			item.Or(
				item.NameContainsFold(q.Search),
				item.DescriptionContainsFold(q.Search),
			),
		)
	}

	if q.Page != -1 || q.PageSize != -1 {
		qb = qb.
			Offset(calculateOffset(q.Page, q.PageSize)).
			Limit(q.PageSize)
	}

	items, err := mapItemsSummaryErr(
		qb.Order(ent.Asc(item.FieldName)).
			WithLabel().
			WithLocation().
			All(ctx),
	)
	if err != nil {
		return PaginationResult[ItemSummary]{}, err
	}

	count, err := qb.Count(ctx)
	if err != nil {
		return PaginationResult[ItemSummary]{}, err
	}

	return PaginationResult[ItemSummary]{
		Page:     q.Page,
		PageSize: q.PageSize,
		Total:    count,
		Items:    items,
	}, nil

}

// GetAll returns all the items in the database with the Labels and Locations eager loaded.
func (e *ItemsRepository) GetAll(ctx context.Context, gid uuid.UUID) ([]ItemSummary, error) {
	return mapItemsSummaryErr(e.db.Item.Query().
		Where(item.HasGroupWith(group.ID(gid))).
		WithLabel().
		WithLocation().
		All(ctx))
}

func (e *ItemsRepository) Create(ctx context.Context, gid uuid.UUID, data ItemCreate) (ItemOut, error) {
	q := e.db.Item.Create().
		SetName(data.Name).
		SetDescription(data.Description).
		SetGroupID(gid).
		SetLocationID(data.LocationID)

	if data.LabelIDs != nil && len(data.LabelIDs) > 0 {
		q.AddLabelIDs(data.LabelIDs...)
	}

	result, err := q.Save(ctx)
	if err != nil {
		return ItemOut{}, err
	}

	return e.GetOne(ctx, result.ID)
}

func (e *ItemsRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return e.db.Item.DeleteOneID(id).Exec(ctx)
}

func (e *ItemsRepository) DeleteByGroup(ctx context.Context, gid, id uuid.UUID) error {
	_, err := e.db.Item.
		Delete().
		Where(
			item.ID(id),
			item.HasGroupWith(group.ID(gid)),
		).Exec(ctx)
	return err
}

func (e *ItemsRepository) UpdateByGroup(ctx context.Context, gid uuid.UUID, data ItemUpdate) (ItemOut, error) {
	q := e.db.Item.Update().Where(item.ID(data.ID), item.HasGroupWith(group.ID(gid))).
		SetName(data.Name).
		SetDescription(data.Description).
		SetLocationID(data.LocationID).
		SetSerialNumber(data.SerialNumber).
		SetModelNumber(data.ModelNumber).
		SetManufacturer(data.Manufacturer).
		SetPurchaseTime(data.PurchaseTime).
		SetPurchaseFrom(data.PurchaseFrom).
		SetPurchasePrice(data.PurchasePrice).
		SetSoldTime(data.SoldTime).
		SetSoldTo(data.SoldTo).
		SetSoldPrice(data.SoldPrice).
		SetSoldNotes(data.SoldNotes).
		SetNotes(data.Notes).
		SetLifetimeWarranty(data.LifetimeWarranty).
		SetInsured(data.Insured).
		SetWarrantyExpires(data.WarrantyExpires).
		SetWarrantyDetails(data.WarrantyDetails).
		SetQuantity(data.Quantity)

	currentLabels, err := e.db.Item.Query().Where(item.ID(data.ID)).QueryLabel().All(ctx)
	if err != nil {
		return ItemOut{}, err
	}

	set := newIDSet(currentLabels)

	for _, l := range data.LabelIDs {
		if set.Contains(l) {
			set.Remove(l)
			continue
		}
		q.AddLabelIDs(l)
	}

	if set.Len() > 0 {
		q.RemoveLabelIDs(set.Slice()...)
	}

	err = q.Exec(ctx)
	if err != nil {
		return ItemOut{}, err
	}

	return e.GetOne(ctx, data.ID)
}
