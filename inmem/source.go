package inmem

import (
	"context"
	"fmt"

	platform "github.com/influxdata/influxdb"
)

// DefaultSource is the default source.
var DefaultSource = platform.Source{
	Default: true,
	Name:    "autogen",
	Type:    platform.SelfSourceType,
}

const (
	// DefaultSourceID it the default source identifier
	DefaultSourceID = "020f755c3c082000"
	// DefaultSourceOrganizationID is the default source's organization identifier
	DefaultSourceOrganizationID = "50616e67652c206c"
)

func init() {
	if err := DefaultSource.ID.DecodeFromString(DefaultSourceID); err != nil {
		panic(fmt.Sprintf("failed to decode default source id: %v", err))
	}

	if err := DefaultSource.OrganizationID.DecodeFromString(DefaultSourceOrganizationID); err != nil {
		panic(fmt.Sprintf("failed to decode default source organization id: %v", err))
	}
}

func (s *Service) initializeSources(ctx context.Context) error {
	_, pe := s.FindSourceByID(ctx, DefaultSource.ID)
	if pe != nil && platform.ErrorCode(pe) != platform.ENotFound {
		return pe
	}

	if platform.ErrorCode(pe) == platform.ENotFound {
		if err := s.PutSource(ctx, &DefaultSource); err != nil {
			return err
		}
	}

	return nil
}

// DefaultSource retrieves the default source.
func (s *Service) DefaultSource(ctx context.Context) (*platform.Source, error) {
	// TODO(desa): make this faster by putting the default source in an index.
	srcs, _, err := s.FindSources(ctx, platform.FindOptions{})
	if err != nil {
		return nil, err
	}

	for _, src := range srcs {
		if src.Default {
			return src, nil
		}
	}
	return nil, &platform.Error{
		Code: platform.ENotFound,
		Msg:  "no default source found",
	}

}

// FindSourceByID retrieves a source by id.
func (s *Service) FindSourceByID(ctx context.Context, id platform.ID) (*platform.Source, error) {
	i, ok := s.sourceKV.Load(id.String())
	if !ok {
		return nil, &platform.Error{
			Code: platform.ENotFound,
			Msg:  platform.ErrSourceNotFound,
		}
	}

	src, ok := i.(*platform.Source)
	if !ok {
		return nil, &platform.Error{
			Code: platform.EInvalid,
			Msg:  fmt.Sprintf("type %T is not a source", i),
		}
	}
	return src, nil
}

// FindSources retrives all sources that match an arbitrary source filter.
// Filters using ID, or OrganizationID and source Name should be efficient.
// Other filters will do a linear scan across all sources searching for a match.
func (s *Service) FindSources(ctx context.Context, opt platform.FindOptions) ([]*platform.Source, int, error) {
	var ds []*platform.Source
	s.sourceKV.Range(func(k, v interface{}) bool {
		d, ok := v.(*platform.Source)
		if !ok {
			return false
		}
		ds = append(ds, d)
		return true
	})
	return ds, len(ds), nil
}

// CreateSource creates a platform source and sets s.ID.
func (s *Service) CreateSource(ctx context.Context, src *platform.Source) error {
	src.ID = s.IDGenerator.ID()
	if err := s.PutSource(ctx, src); err != nil {
		return &platform.Error{
			Err: err,
		}
	}
	return nil
}

// PutSource will put a source without setting an ID.
func (s *Service) PutSource(ctx context.Context, src *platform.Source) error {
	s.sourceKV.Store(src.ID.String(), src)
	return nil
}

// UpdateSource updates a source according the parameters set on upd.
func (s *Service) UpdateSource(ctx context.Context, id platform.ID, upd platform.SourceUpdate) (*platform.Source, error) {
	src, err := s.FindSourceByID(ctx, id)
	if err != nil {
		return nil, &platform.Error{
			Err: err,
			Op:  OpPrefix + platform.OpUpdateSource,
		}
	}

	upd.Apply(src)
	s.sourceKV.Store(src.ID.String(), src)
	return src, nil
}

// DeleteSource deletes a source and prunes it from the index.
func (s *Service) DeleteSource(ctx context.Context, id platform.ID) error {
	if _, err := s.FindSourceByID(ctx, id); err != nil {
		return &platform.Error{
			Err: err,
			Op:  OpPrefix + platform.OpDeleteSource,
		}
	}
	s.sourceKV.Delete(id.String())
	return nil
}
