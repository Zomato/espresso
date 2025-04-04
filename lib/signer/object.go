package signer

import (
	"bytes"
	"fmt"

	cContext "context"

	log "github.com/Zomato/espresso/lib/logger"
)

func (context *SignContext) addObject(object []byte) (uint32, error) {
	ctx := cContext.Background()

	if context.lastXrefID == 0 {
		lastXrefID, err := context.getLastObjectIDFromXref()
		if err != nil {
			log.Logger.Error(ctx, "failed to get last object ID", err, nil)
			return 0, fmt.Errorf("failed to get last object ID: %w", err)
		}
		context.lastXrefID = lastXrefID
	}

	objectID := context.lastXrefID + uint32(len(context.newXrefEntries)) + 1
	context.newXrefEntries = append(context.newXrefEntries, xrefEntry{
		ID:     objectID,
		Offset: int64(context.OutputBuffer.Buff.Len()) + 1,
	})

	err := context.writeObject(objectID, object)
	if err != nil {
		log.Logger.Error(ctx, "failed to write object", err, nil)
		return 0, fmt.Errorf("failed to write object: %w", err)
	}

	return objectID, nil
}

func (context *SignContext) updateObject(id uint32, object []byte) error {
	context.updatedXrefEntries = append(context.updatedXrefEntries, xrefEntry{
		ID:     id,
		Offset: int64(context.OutputBuffer.Buff.Len()) + 1,
	})

	err := context.writeObject(id, object)
	if err != nil {
		log.Logger.Error(cContext.Background(), "failed to write object", err, nil)
		return fmt.Errorf("failed to write object: %w", err)
	}

	return nil
}

func (context *SignContext) writeObject(id uint32, object []byte) error {
	ctx := cContext.Background()

	if _, err := context.OutputBuffer.Write([]byte(fmt.Sprintf("\n%d 0 obj\n", id))); err != nil {
		log.Logger.Error(ctx, "failed to write object header", err, nil)
		return fmt.Errorf("failed to write object header: %w", err)
	}

	object = bytes.TrimSpace(object)
	if _, err := context.OutputBuffer.Write(object); err != nil {
		log.Logger.Error(ctx, "failed to write object content", err, nil)
		return fmt.Errorf("failed to write object content: %w", err)
	}

	if _, err := context.OutputBuffer.Write([]byte(objectFooter)); err != nil {
		log.Logger.Error(ctx, "failed to write object footer", err, nil)
		return fmt.Errorf("failed to write object footer: %w", err)
	}

	return nil
}
