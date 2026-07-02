package opcua

import (
	"context"
	"fmt"

	"github.com/gopcua/opcua/ua"

	"pharma-platform/internal/models"
)

func (c *Client) readTag(
	ctx context.Context,
	tag models.Tag,
) (models.Sample, error) {

	nodeID, err := ua.ParseNodeID(tag.Address)
	if err != nil {
		return models.Sample{}, fmt.Errorf("parse node id: %w", err)
	}

	req := &ua.ReadRequest{
		MaxAge:             0,
		TimestampsToReturn: ua.TimestampsToReturnBoth,
		NodesToRead: []*ua.ReadValueID{
			{
				NodeID:      nodeID,
				AttributeID: ua.AttributeIDValue,
			},
		},
	}

	resp, err := c.client.Read(ctx, req)
	if err != nil {
		return models.Sample{}, fmt.Errorf("read node: %w", err)
	}

	if len(resp.Results) != 1 {
		return models.Sample{}, fmt.Errorf("expected 1 result, got %d", len(resp.Results))
	}

	result := resp.Results[0]

	return models.Sample{
		Timestamp: result.SourceTimestamp,
		PLCID:     tag.PLCID,
		TagID:     tag.ID,
		Value:     result.Value.Value(),
		Quality:   QualityFromStatus(result.Status),
	}, nil
}
