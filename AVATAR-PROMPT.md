# Telegram bot avatar — ChatGPT image prompt

Use in ChatGPT (or DALL·E) → generate **square 1:1**, then crop to circle for BotFather (`/setuserpic`).

## Primary prompt (EN)

```
App icon for a Telegram support bot named DocPilot. Square 1:1, centered subject, clean edges for circular crop.

Visual: a minimal paper airplane folded from a sheet of documentation, flying through a soft sage-green and warm parchment atmosphere. Tiny citation marks / quotation ticks as subtle geometric accents around the plane. Modern SaaS product icon, flat-to-soft 3D, high clarity at small size.

Style: refined product icon, restrained palette (sage green #0F7A5F, parchment #F3EFE6, deep ink #14201C), no purple neon, no glow spam, no photorealistic face, no text, no logo letters, no watermark. Soft studio lighting, generous negative space, crisp silhouette.
```

## Variants (if first pass is weak)

**A — abstract mark**
```
Minimal geometric mark: overlapping rounded rectangle (document) and a small chevron arrow (send/reply). Sage green on parchment. Flat vector app icon, square, no text, sharp for 160×160 Telegram avatar.
```

**B — library spark**
```
Single open book seen from above turning into a soft spiral of light lines suggesting retrieval/search. Warm parchment background, sage accent lines, ink outlines. Clean SaaS icon, square 1:1, no letters, no people.
```

## BotFather tips

1. Prefer **512×512** PNG, then Telegram will round it.
2. Subject should sit in the **center 70%** — edges get clipped by the circle.
3. Avoid thin lines; they disappear at chat-list size.
4. After upload: `/setuserpic` → `/setdescription` → `/setabouttext`.

## Suggested BotFather copy

**Name:** DocPilot Support  
**About:** Answers from your docs with sources. Tap /human for a person.  
**Description:** Docs-grounded AI support bot. Ask about refunds, onboarding, or SLA — or escalate to a human anytime.
