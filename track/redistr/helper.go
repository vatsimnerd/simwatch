package redistr

import (
	"context"
	"strconv"
)

const (
	keyTrackIDs    = "track_ids"
	keyPrefixTrack = "tracks"
	keySeparator   = ":"
)

func pointsIndexKey(trackID string) string {
	return keyPrefixTrack +
		keySeparator +
		trackID +
		keySeparator +
		"points"
}

func pointKey(trackID string, ts string) string {
	return pointsIndexKey(trackID) + keySeparator + ts
}

func trackCreatedKey(trackID string) string {
	return keyPrefixTrack + keySeparator + trackID + keySeparator + "created_at"
}

func (r *RedisReadWriter) getHInt(ctx context.Context, key string, field string) (int, error) {
	sv, err := r.cli.HGet(ctx, key, field).Result()
	if err != nil {
		return 0, err
	}
	iv, err := strconv.ParseInt(sv, 10, 64)
	if err != nil {
		return 0, err
	}
	return int(iv), nil
}

func (r *RedisReadWriter) getHFloat(ctx context.Context, key string, field string) (float64, error) {
	sv, err := r.cli.HGet(ctx, key, field).Result()
	if err != nil {
		return 0, err
	}
	return strconv.ParseFloat(sv, 64)
}

func (r *RedisReadWriter) getInt64(ctx context.Context, key string) (int64, error) {
	sv, err := r.cli.Get(ctx, key).Result()
	if err != nil {
		return 0, err
	}
	return strconv.ParseInt(sv, 10, 64)
}
