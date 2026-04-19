INSERT INTO beer_features (beer_id, feature_id)
VALUES ($1, $2)
ON CONFLICT DO NOTHING;