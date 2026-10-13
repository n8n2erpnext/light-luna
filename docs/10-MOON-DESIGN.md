# 10 — Light Luna Brand & Moon-State System

## Identity

**Brand:** Light Luna (two words, title case). **Family:** Light Remote, LightBI, Light Luna; by THAI DUY Digital. **Tagline candidate:** *Orbit. Observe. Defend.* **Long form:** *Intelligent Firewall & Autonomous Defense.*

## Core symbol — abstract orbital crescent

Default icon: **technology-abstract crescent / new moon silhouette**, navy deep-space background, moonlight silver and electric cyan. Orbital stroke subtly intersects or wraps the moon; a single pulse node expresses patrol/intelligence. At a second look, viewers discover the crescent; avoid literal frog, bulky padlock, shield or dense crater texture as mandatory motifs. Ensure distinct at 16–24 px and in 1-color emboss/laser-engraving for future appliances. Brand mark remains stable; status-icon variants change meaningfully.

## States

| Semantic state | Icon/palette | Trigger / explanation | Caution |
|---|---|---|---|
| `MOON_CALM` | Cyan/new crescent | Local sensors current; no actionable threat in observation window | Never claim total safety |
| `MOON_WATCH` | Gold/yellow crescent | Suspicion requiring more evidence | Not proof of attack |
| `MOON_ALERT` | Amber/orange moon | Elevated verified risk or reversible mitigation proposed | Label current action separately |
| `BLOOD_MOON` | Deep red/crimson moon | Confirmed or strongly evidenced serious active threat | Real ban may or may not be active |
| `ECLIPSE` | Dim, occluded moon, uncertainty ring | Missing/stale/untrusted sensor or management visibility | Not a numerical threat rank; can coexist with incident severity |

Moon phases are **semantic UI motifs, not astronomical event forecasts or a direct scale of danger**. Severe threat and stale telemetry are orthogonal; display both dimensions, e.g. `BLOOD_MOON + ECLIPSE` when a last-known serious incident coexists with telemetry loss.

## Suggested tokens (design proposals)

`luna.night #0B1728`, `luna.cyan #67DCEB`, `luna.moon #DDEAF4`, `luna.watch #E8BE54`, `luna.alert #E88B40`, `luna.blood #C83F4B`, `luna.eclipse #78869C`. Check WCAG color contrast in both light/dark modes; no status represented by hue alone. Avoid animations that distract, and respect reduced motion.

## Immersive metaphor (optional storytelling)

Earth = infrastructure, orbital patrol = repeated monitoring, meteoroids = scan/botnet, craters = verified incident memories, tidal rhythm = system health and balance. Scientifically, the Moon is not an asteroid shield that captures most Earth-bound objects; say “observes impacts and retains their history,” not “physically prevents most impacts.” Keep copy professional on safety-critical pages.

## UI variants

- Stable product logo in navigation/login/legal docs.
- Dynamic moon-state in tray, device card, incident feed, WebUI title badge and push notifications.
- Optional moon phase as non-critical ambient design; functional threat-state always has text + timestamp + current telemetry quality.
- Hardware appliance: monochrome icon/wordmark first; status LED does not substitute screen/CLI status explanation.

## Future work

Create actual vector master in Figma with icon grid, clear-space, safe-area, light/dark variants, status-token accessibility testing, 16px recognizability and original/usage trademark review. **No final logo files exist in this baseline.**
