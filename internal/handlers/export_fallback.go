package handlers

import (
	"fmt"
	"strings"
)

// ── Fallback brand palette ──────────────────────────────────────────────────
// Mirrors the frontend's extractPalette() in VisualIdentityPanel.tsx so the
// exported fallback assets use the same colours the user saw on screen.

type fallbackPalette struct {
	Bg      string
	Accent  string
	Accent2 string
	Text    string
}

func exportPalette(colorMood string) fallbackPalette {
	lower := strings.ToLower(colorMood)

	bg := "#0a1628"
	switch {
	case strings.Contains(lower, "midnight"), strings.Contains(lower, "void"), strings.Contains(lower, "black"):
		bg = "#050810"
	case strings.Contains(lower, "dark blue"), strings.Contains(lower, "navy"):
		bg = "#0a1628"
	case strings.Contains(lower, "dark"):
		bg = "#0d0d1a"
	}

	type rule struct{ key, hex string }
	rules := []rule{
		{"bright orange", "#f97316"}, {"orange", "#f97316"}, {"coral", "#f97316"},
		{"sky blue", "#38bdf8"}, {"sky", "#38bdf8"}, {"azure", "#38bdf8"}, {"cerulean", "#38bdf8"},
		{"electric blue", "#3b82f6"}, {"electric", "#3b82f6"}, {"neon", "#3b82f6"},
		{"yellow", "#facc15"}, {"gold", "#facc15"}, {"amber", "#facc15"},
		{"teal", "#22d3ee"}, {"cyan", "#22d3ee"}, {"aqua", "#22d3ee"},
		{"green", "#10b981"}, {"emerald", "#10b981"}, {"lime", "#10b981"},
		{"purple", "#8B5CF6"}, {"violet", "#8B5CF6"}, {"indigo", "#8B5CF6"},
		{"pink", "#ec4899"}, {"rose", "#ec4899"}, {"magenta", "#ec4899"},
		{"red", "#ef4444"}, {"crimson", "#ef4444"}, {"scarlet", "#ef4444"},
		{"blue", "#3b82f6"},
	}

	var found []string
	for _, r := range rules {
		if strings.Contains(lower, r.key) {
			dup := false
			for _, f := range found {
				if f == r.hex {
					dup = true
					break
				}
			}
			if !dup {
				found = append(found, r.hex)
			}
		}
		if len(found) >= 2 {
			break
		}
	}
	accent, accent2 := "#f97316", "#38bdf8"
	if len(found) > 0 {
		accent = found[0]
	}
	if len(found) > 1 {
		accent2 = found[1]
	}

	return fallbackPalette{Bg: bg, Accent: accent, Accent2: accent2, Text: "#ffffff"}
}

// initials returns up to 2 uppercase letters for a brand name, matching the
// frontend's initials() logic: first letter of first two words, else first
// two characters of a single word.
func initials(name string) string {
	parts := strings.Fields(name)
	if len(parts) >= 2 {
		return strings.ToUpper(string([]rune(parts[0])[0]) + string([]rune(parts[1])[0]))
	}
	r := []rune(name)
	if len(r) >= 2 {
		return strings.ToUpper(string(r[:2]))
	}
	if len(r) == 1 {
		return strings.ToUpper(string(r))
	}
	return "NV"
}

func firstInitial(name string) string {
	r := []rune(strings.TrimSpace(name))
	if len(r) == 0 {
		return "N"
	}
	return strings.ToUpper(string(r[0]))
}

const svgFontStack = `'Space Grotesk', Arial, Helvetica, sans-serif`

// hexPolygonPoints returns the 6-point hexagon used throughout the frontend's
// CSS clip-path marks, scaled to fit a box of the given size.
func hexPolygonPoints(size float64) string {
	s := size
	return fmt.Sprintf("%.1f,0 %.1f,%.1f %.1f,%.1f %.1f,%.1f 0,%.1f 0,%.1f",
		s/2, s, s*0.25, s, s*0.75, s/2, s, s*0.75, s*0.25)
}

// ── Fallback logo generators ────────────────────────────────────────────────
// Each returns a complete, standalone SVG document string.

func fallbackLogoProfileSVG(name string, pal fallbackPalette) string {
	return fmt.Sprintf(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 200 200">
  <rect width="200" height="200" fill="%s"/>
  <circle cx="100" cy="100" r="88" fill="%s"/>
  <text x="100" y="100" font-family="%s" font-size="88" font-weight="900"
    fill="%s" text-anchor="middle" dominant-baseline="central">%s</text>
</svg>`, pal.Bg, pal.Accent, svgFontStack, pal.Bg, escXML(firstInitial(name)))
}

func fallbackLogoAppSVG(name string, pal fallbackPalette) string {
	return fmt.Sprintf(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 200 200">
  <defs>
    <linearGradient id="g" x1="0%%" y1="0%%" x2="100%%" y2="100%%">
      <stop offset="0%%" stop-color="%s"/>
      <stop offset="100%%" stop-color="%s"/>
    </linearGradient>
  </defs>
  <rect width="200" height="200" fill="%s"/>
  <rect x="18" y="18" width="164" height="164" rx="38" fill="url(#g)"/>
  <text x="100" y="104" font-family="%s" font-size="62" font-weight="900"
    fill="%s" text-anchor="middle" dominant-baseline="central">%s</text>
</svg>`, pal.Accent, pal.Accent2, pal.Bg, svgFontStack, pal.Bg, escXML(initials(name)))
}

func fallbackLogoBusinessSVG(name, tagline string, pal fallbackPalette) string {
	hexPts := hexPolygonPoints(36)
	tag := strings.ToUpper(tagline)
	if tag == "" {
		tag = "BRAND IDENTITY"
	}
	return fmt.Sprintf(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 400 200">
  <defs>
    <linearGradient id="g" x1="0%%" y1="0%%" x2="100%%" y2="100%%">
      <stop offset="0%%" stop-color="%s"/>
      <stop offset="100%%" stop-color="%s"/>
    </linearGradient>
  </defs>
  <rect width="400" height="200" fill="%s"/>
  <rect x="36" y="58" width="328" height="84" rx="14" fill="rgba(255,255,255,0.05)"
    stroke="%s" stroke-opacity="0.35" stroke-width="1"/>
  <g transform="translate(58,82)"><polygon points="%s" fill="url(#g)"/></g>
  <text x="112" y="102" font-family="%s" font-size="30" font-weight="900"
    fill="#ffffff">%s</text>
  <text x="112" y="126" font-family="%s" font-size="11" font-weight="700"
    letter-spacing="2" fill="%s">%s</text>
</svg>`, pal.Accent, pal.Accent2, pal.Bg, pal.Accent, hexPts,
		svgFontStack, escXML(name), svgFontStack, pal.Accent, escXML(tag))
}

// ── Fallback mood board tiles ────────────────────────────────────────────────
// 4 panels mirroring the frontend's CSSMoodBoardFallback tile themes.

func fallbackMoodBoardSVGs(name, tagline, personality string, pal fallbackPalette) []string {
	hexPts28 := hexPolygonPoints(28)

	tile0 := fmt.Sprintf(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 300 300">
  <rect width="300" height="300" fill="%s"/>
  <circle cx="150" cy="130" r="80" fill="%s" opacity="0.35"/>
  <text x="150" y="220" font-family="%s" font-size="12" font-weight="900"
    letter-spacing="2" fill="%s" text-anchor="middle">COLOUR WORLD</text>
  <circle cx="118" cy="252" r="9" fill="%s"/>
  <circle cx="140" cy="252" r="9" fill="%s"/>
  <circle cx="162" cy="252" r="9" fill="%s"/>
  <circle cx="184" cy="252" r="9" fill="#ffffff"/>
</svg>`, pal.Bg, pal.Accent, svgFontStack, pal.Accent, pal.Bg, pal.Accent, pal.Accent2)

	tile1 := fmt.Sprintf(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 300 300">
  <rect width="300" height="300" fill="%s"/>
  <g transform="translate(134,90)"><polygon points="%s" fill="url(#g1)"/></g>
  <defs><linearGradient id="g1" x1="0%%" y1="0%%" x2="100%%" y2="100%%">
    <stop offset="0%%" stop-color="%s"/><stop offset="100%%" stop-color="%s"/>
  </linearGradient></defs>
  <text x="150" y="170" font-family="%s" font-size="26" font-weight="900"
    fill="#ffffff" text-anchor="middle">%s</text>
  <text x="150" y="196" font-family="%s" font-size="13" font-style="italic"
    fill="%s" text-anchor="middle">&#8220;%s&#8221;</text>
  <text x="150" y="222" font-family="%s" font-size="11" fill="%s" opacity="0.6" text-anchor="middle">%s</text>
</svg>`, pal.Bg, hexPts28, pal.Accent, pal.Accent2, svgFontStack, escXML(name),
		svgFontStack, pal.Accent, escXML(tagline), svgFontStack, pal.Text, escXML(orDefault(personality, "modern brand")))

	tile2 := buildPatternTile(pal)

	tile3 := fmt.Sprintf(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 300 300">
  <rect width="300" height="300" fill="%s"/>
  <text x="24" y="40" font-family="%s" font-size="11" font-weight="900"
    letter-spacing="2" fill="%s">TYPOGRAPHY</text>
  <text x="24" y="110" font-family="%s" font-size="64" font-weight="900" fill="#ffffff">Aa</text>
  <text x="24" y="134" font-family="%s" font-size="11" font-weight="700" fill="%s">SPACE GROTESK</text>
  <rect x="24" y="248" width="110" height="30" rx="4" fill="%s"/>
  <text x="79" y="268" font-family="%s" font-size="11" font-weight="900"
    fill="%s" text-anchor="middle">CTA Button</text>
</svg>`, pal.Bg, svgFontStack, pal.Accent, svgFontStack, svgFontStack, pal.Text, pal.Accent, svgFontStack, pal.Bg)

	return []string{tile0, tile1, tile2, tile3}
}

func buildPatternTile(pal fallbackPalette) string {
	var shapes strings.Builder
	hexPts := hexPolygonPoints(24)
	for i := 0; i < 16; i++ {
		col := i % 4
		row := i / 4
		x := float64(col)*72 + 8
		y := float64(row)*72 + 16
		fill := pal.Accent
		if i%2 != 0 {
			fill = pal.Accent2
		}
		opacity := "0.55"
		if i%3 == 0 {
			opacity = "0.30"
		}
		rot := i * 22
		fmt.Fprintf(&shapes,
			`<g transform="translate(%.1f,%.1f) rotate(%d,12,12)"><polygon points="%s" fill="%s" opacity="%s"/></g>`,
			x, y, rot, hexPts, fill, opacity)
	}
	return fmt.Sprintf(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 300 300">
  <rect width="300" height="300" fill="%s"/>
  %s
  <text x="16" y="288" font-family="%s" font-size="11" font-weight="900"
    letter-spacing="2" fill="%s">PATTERN DNA</text>
</svg>`, pal.Bg, shapes.String(), svgFontStack, pal.Accent)
}

// ── Fallback landing page ───────────────────────────────────────────────────

func fallbackLandingPageHTML(req exportRequest, pal fallbackPalette) string {
	name := req.BrandName
	tagline := req.Card.Tagline
	desc := req.Card.ShortDesc
	industry := strings.ToUpper(orDefault(req.Intake.Industry, "Creative"))
	audience := orDefault(req.Intake.TargetAudience, "Everyone")
	personality := orDefault(req.Intake.Personality, "Creative")

	var logoSVG string
	switch req.SelectedLogoKey {
	case "app":
		logoSVG = fallbackLogoAppSVG(name, pal)
	case "profile":
		logoSVG = fallbackLogoProfileSVG(name, pal)
	default:
		logoSVG = fallbackLogoBusinessSVG(name, tagline, pal)
	}

	var b strings.Builder
	b.WriteString("<!DOCTYPE html>\n<html lang=\"en\">\n<head>\n")
	b.WriteString("<meta charset=\"UTF-8\">\n<meta name=\"viewport\" content=\"width=device-width,initial-scale=1\">\n")
	fmt.Fprintf(&b, "<title>%s</title>\n", escHTML(name))
	b.WriteString("<style>\n")
	b.WriteString("*{box-sizing:border-box;margin:0;padding:0;}\n")
	fmt.Fprintf(&b, "body{font-family:-apple-system,BlinkMacSystemFont,\"Segoe UI\",sans-serif;background:linear-gradient(150deg,%s 0%%,%s 55%%,%s22 100%%);color:#fff;min-height:100vh;}\n",
		pal.Bg, pal.Bg, pal.Accent)
	b.WriteString("nav{display:flex;align-items:center;justify-content:space-between;padding:18px 32px;border-bottom:1px solid rgba(255,255,255,0.08);}\n")
	b.WriteString(".brand{display:flex;align-items:center;gap:10px;font-weight:900;font-size:1.1rem;}\n")
	b.WriteString(".brand .mark{width:22px;height:22px;flex-shrink:0;}\n")
	fmt.Fprintf(&b, ".navlinks{display:flex;gap:20px;align-items:center;font-size:0.85rem;font-weight:700;color:%s;}\n", pal.Accent2)
	fmt.Fprintf(&b, ".navlinks .cta{background:%s;color:%s;padding:7px 16px;border-radius:8px;}\n", pal.Accent, pal.Bg)
	b.WriteString(".hero{display:flex;flex-wrap:wrap;align-items:center;gap:40px;padding:64px 32px;max-width:1100px;margin:0 auto;}\n")
	b.WriteString(".hero-text{flex:1 1 420px;}\n")
	fmt.Fprintf(&b, ".eyebrow{font-size:0.8rem;font-weight:900;letter-spacing:2px;color:%s;margin-bottom:10px;}\n", pal.Accent)
	b.WriteString("h1{font-size:3rem;font-weight:900;letter-spacing:-1px;margin-bottom:14px;}\n")
	fmt.Fprintf(&b, ".rule{width:52px;height:4px;background:%s;border-radius:2px;margin-bottom:14px;}\n", pal.Accent)
	fmt.Fprintf(&b, ".tagline{font-size:1.1rem;font-style:italic;color:%s;margin-bottom:14px;}\n", pal.Accent2)
	b.WriteString(".desc{color:rgba(255,255,255,0.7);margin-bottom:26px;max-width:480px;}\n")
	b.WriteString(".btnrow{display:flex;gap:12px;flex-wrap:wrap;}\n")
	fmt.Fprintf(&b, ".btn-primary{background:%s;color:%s;font-weight:900;padding:12px 26px;border-radius:10px;text-decoration:none;}\n",
		pal.Accent, pal.Bg)
	fmt.Fprintf(&b, ".btn-secondary{border:1px solid %s;color:#fff;font-weight:700;padding:12px 22px;border-radius:10px;text-decoration:none;}\n",
		pal.Accent2)
	b.WriteString(".hero-visual{flex:0 0 220px;display:flex;justify-content:center;}\n")
	b.WriteString(".hero-visual svg{width:100%;max-width:220px;height:auto;}\n")
	b.WriteString(".features{display:flex;border-top:1px solid rgba(255,255,255,0.1);}\n")
	b.WriteString(".features div{flex:1;text-align:center;padding:16px 8px;font-size:0.75rem;font-weight:900;letter-spacing:1px;color:rgba(255,255,255,0.65);border-right:1px solid rgba(255,255,255,0.06);}\n")
	b.WriteString(".features div:last-child{border-right:none;}\n")
	b.WriteString("</style>\n</head>\n<body>\n")

	b.WriteString("  <nav>\n    <div class=\"brand\">")
	fmt.Fprintf(&b, "%s", escHTML(name))
	b.WriteString("</div>\n    <div class=\"navlinks\">\n      <span>About</span><span>Features</span><span class=\"cta\">Get Started</span>\n    </div>\n  </nav>\n")

	b.WriteString("  <div class=\"hero\">\n    <div class=\"hero-text\">\n")
	fmt.Fprintf(&b, "      <div class=\"eyebrow\">%s</div>\n", escHTML(industry))
	fmt.Fprintf(&b, "      <h1>%s</h1>\n", escHTML(name))
	b.WriteString("      <div class=\"rule\"></div>\n")
	fmt.Fprintf(&b, "      <div class=\"tagline\">&#8220;%s&#8221;</div>\n", escHTML(tagline))
	fmt.Fprintf(&b, "      <p class=\"desc\">%s</p>\n", escHTML(desc))
	b.WriteString("      <div class=\"btnrow\">\n")
	fmt.Fprintf(&b, "        <a class=\"btn-primary\" href=\"#\">Start with %s &rarr;</a>\n", escHTML(name))
	b.WriteString("        <a class=\"btn-secondary\" href=\"#\">Learn More</a>\n")
	b.WriteString("      </div>\n    </div>\n")
	fmt.Fprintf(&b, "    <div class=\"hero-visual\">%s</div>\n", logoSVG)
	b.WriteString("  </div>\n")

	b.WriteString("  <div class=\"features\">\n")
	fmt.Fprintf(&b, "    <div>%s</div>\n", escHTML(industry))
	fmt.Fprintf(&b, "    <div>%s</div>\n", escHTML(personality))
	fmt.Fprintf(&b, "    <div>%s</div>\n", escHTML(audience))
	b.WriteString("  </div>\n</body>\n</html>")

	return b.String()
}

func orDefault(s, def string) string {
	if strings.TrimSpace(s) == "" {
		return def
	}
	return s
}

// escXML is an alias of escHTML — same escaping rules apply inside SVG text nodes.
func escXML(s string) string { return escHTML(s) }
