# Feature: Collaborative Annotations

Allow engineers to annotate charts, recipes, and wafers with comments and share them with the team.

## User Story
As a process engineer, I want to mark up an OES chart with notes about what I observed and share it with my colleague.

## Implementation Details

### Canvas Interaction
- **Spatial Metadata**: Annotations are not just pixels; they are linked to specific (time, wavelength, intensity) coordinates in the OES data.
- **Rich Media Support**: Attach SEM images or Excel reports directly to a point on a trend line.

### Backend: Collaboration Service
```typescript
// services/collaboration/annotations.ts

interface Annotation {
  targetId: string; // e.g., run_id or recipe_id
  coordinates: { x: number; y: number; field: string };
  content: string;
  authorId: string;
  mentions: string[];
}

export class AnnotationManager {
  async save(note: Annotation) {
    // 1. Persist to DB
    await db.annotations.create(note);
    
    // 2. Notify mentioned users
    if (note.mentions.length > 0) {
      await notificationService.sendMentions(note.mentions, note);
    }
  }
}
```

## Killer Differentiator: Data-Centric Discussion
- **Context is King**: Comments aren't stuck in email; they live exactly where the data issue happened.
- **Multi-User Sync**: Real-time "Google Docs style" collaboration where engineers can see each other's highlights during a remote meeting.
- **Permanent Audit Trail**: All engineering decisions regarding a recipe shift are documented and signed off within the tool.
