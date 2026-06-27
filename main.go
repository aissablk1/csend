// csend — inter-session messaging for CLI coding agents (Claude Code & co.),
// state-aware and provider-agnostic, driven through the cmux / Vibe Island
// control socket. MVP slice 1: cmux backend + Claude adapter + guarded delivery.
package main

import (
	"fmt"
	"os"
	"sort"
	"strings"
	"text/tabwriter"
)

const usage = `csend — messagerie inter-sessions pour agents CLI (Claude Code & co.)

Usage:
  csend list                       liste les sessions agent + leur état
  csend tree                       affiche le graphe familial (père/enfants) des sessions
  csend send <cible> <message…>    envoie un message (auto: valide si la cible est idle)
        --stage                    dépose le texte sans valider (pas d'Entrée)
        --send                     valide même si la cible est occupée
        --force                    passe outre tous les garde-fous (dangereux)
  csend send --down <message…>     broadcast vers les ENFANTS de la session courante
  csend send --up <message…>       envoie au PARENT de la session courante
        --to-siblings              vers les frères ; --to-descendants vers tout le sous-arbre
        --from <session>           change la base (défaut: la session courante)
  csend read <cible> [--lines N]   lit l'écran d'une session
  csend recv [--peek]              lit (et vide) l'inbox coopératif de CETTE session
  csend inbox <cible> <message…>   dépose un message coopératif (hors cmux, tout OS)
  csend id [--create]              affiche/crée l'identité crypto locale (vault chiffré)
  csend recovery split <K> <N>     découpe l'identité en N parts Shamir (seuil K)
  csend recovery combine <part…>   reconstitue l'identité depuis ≥ K parts
  csend link <enfant> <parent>     déclare <parent> comme parent de <enfant>
  csend unlink <enfant>            détache <enfant> de son parent
  csend help

Cible: nom de workspace (ex. SACEM), session-id (ex. 7f384610), ou ref (surface:42).
Garde-fous: jamais d'auto-injection dans la session courante ; jamais d'Entrée sur
un prompt de confirmation (y/N) ou une session non reconnue (sauf --force).`

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, usage)
		os.Exit(2)
	}
	switch os.Args[1] {
	case "list", "ls":
		mustBackend()
		cmdList()
	case "tree":
		mustBackend()
		cmdTree()
	case "link":
		mustBackend()
		cmdLink(os.Args[2:])
	case "unlink":
		mustBackend()
		cmdUnlink(os.Args[2:])
	case "send":
		mustBackend()
		cmdSend(os.Args[2:])
	case "read":
		mustBackend()
		cmdRead(os.Args[2:])
	case "recv":
		cmdRecv(os.Args[2:]) // cooperative path — no cmux backend required
	case "inbox":
		cmdInbox(os.Args[2:])
	case "id":
		cmdID(os.Args[2:])
	case "recovery":
		cmdRecovery(os.Args[2:])
	case "_why": // hidden diagnostic
		mustBackend()
		tgt, err := resolveTarget(os.Args[2])
		if err != nil {
			fail(err.Error())
		}
		screen, err := readScreen(tgt.Ref, tgt.WorkspaceRef, stateTailLines)
		if err != nil {
			fail(err.Error())
		}
		fmt.Println(StateDebug(screen))
	case "help", "-h", "--help":
		fmt.Println(usage)
	default:
		fmt.Fprintf(os.Stderr, "commande inconnue: %s\n\n%s\n", os.Args[1], usage)
		os.Exit(2)
	}
}

func mustBackend() {
	if !backendAvailable() {
		fail("aucun backend de terminal disponible (cmux ou tmux). Lance l'un des deux, " +
			"ou utilise la voie coopérative : csend inbox / csend recv.")
	}
}

func cmdList() {
	sessions, err := listSessions()
	if err != nil {
		fail(err.Error())
	}
	if len(sessions) == 0 {
		fmt.Println("Aucune session agent détectée.")
		return
	}
	w := tabwriter.NewWriter(os.Stdout, 0, 2, 2, ' ', 0)
	fmt.Fprintln(w, "ÉTAT\tREF\tWORKSPACE\tSESSION\tTITRE")
	for _, s := range sessions {
		here := ""
		if s.Here {
			here = " (ici)"
		}
		fmt.Fprintf(w, "%s\t%s%s\t%s\t%s\t%s\n",
			glyph(s.State), s.Ref, here, s.Workspace, shortID(s.SessionID), truncate(s.Title, 40))
	}
	w.Flush()
}

func cmdTree() {
	sessions, err := listSessions()
	if err != nil {
		fail(err.Error())
	}
	rel := loadRelations()
	byKey := map[string]Session{}
	for _, s := range sessions {
		byKey[sessionKey(s.Surface)] = s
	}
	// Every node = live session ∪ relation endpoint. Roots have no parent edge.
	all := map[string]bool{}
	for k := range byKey {
		all[k] = true
	}
	for _, e := range rel.Edges {
		all[e.Parent] = true
		all[e.Child] = true
	}
	var roots []string
	for id := range all {
		if _, hasParent := rel.parentOf(id); !hasParent {
			roots = append(roots, id)
		}
	}
	sort.Strings(roots)
	if len(roots) == 0 {
		fmt.Println("Aucune session.")
		return
	}
	seen := map[string]bool{}
	var render func(id, prefix string, last bool, depth int)
	render = func(id, prefix string, last bool, depth int) {
		if seen[id] {
			return // cycle guard (also enforced at link time)
		}
		seen[id] = true
		branch, childPrefix := "", prefix
		if depth > 0 {
			if last {
				branch, childPrefix = "└─ ", prefix+"   "
			} else {
				branch, childPrefix = "├─ ", prefix+"│  "
			}
		}
		fmt.Printf("%s%s%s\n", prefix, branch, nodeLabel(id, byKey, rel))
		kids := rel.childrenOf(id)
		sort.Strings(kids)
		for i, k := range kids {
			render(k, childPrefix, i == len(kids)-1, depth+1)
		}
	}
	for i, root := range roots {
		render(root, "", i == len(roots)-1, 0)
	}
}

func nodeLabel(id string, byKey map[string]Session, rel *Relations) string {
	if s, ok := byKey[id]; ok {
		here := ""
		if s.Here {
			here = " (ici)"
		}
		return fmt.Sprintf("%s  %s%s  %s  %s", glyph(s.State), shortKey(id), here, s.Workspace, truncate(s.Title, 34))
	}
	name := rel.Names[id]
	if name == "" {
		name = id
	}
	return fmt.Sprintf("○ offline  %s  %s", shortKey(id), name)
}

func cmdLink(args []string) {
	if len(args) < 2 {
		fail("usage: csend link <enfant> <parent>")
	}
	child, err := resolveTarget(args[0])
	if err != nil {
		fail(err.Error())
	}
	parent, err := resolveTarget(args[1])
	if err != nil {
		fail(err.Error())
	}
	ck, pk := sessionKey(child), sessionKey(parent)
	if ck == pk {
		fail("une session ne peut pas être son propre parent")
	}
	rel := loadRelations()
	if rel.wouldCycle(pk, ck) {
		fail("refusé: créerait un cycle dans l'arbre familial")
	}
	rel.link(pk, ck)
	rel.Names[ck], rel.Names[pk] = child.Workspace, parent.Workspace
	if err := rel.save(); err != nil {
		fail(err.Error())
	}
	fmt.Printf("✓ lié: %s (%s) → enfant de %s (%s)\n", shortKey(ck), child.Workspace, shortKey(pk), parent.Workspace)
}

func cmdUnlink(args []string) {
	if len(args) < 1 {
		fail("usage: csend unlink <enfant>")
	}
	child, err := resolveTarget(args[0])
	if err != nil {
		fail(err.Error())
	}
	ck := sessionKey(child)
	rel := loadRelations()
	if _, ok := rel.parentOf(ck); !ok {
		fmt.Printf("%s n'avait pas de parent.\n", shortKey(ck))
		return
	}
	rel.unlinkChild(ck)
	if err := rel.save(); err != nil {
		fail(err.Error())
	}
	fmt.Printf("✓ détaché: %s (%s)\n", shortKey(ck), child.Workspace)
}

func cmdSend(args []string) {
	mode := ModeAuto
	var dir *Direction
	from := ""
	var pos []string
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--stage":
			mode = ModeStage
		case "--send":
			mode = ModeSend
		case "--force":
			mode = ModeForce
		case "--up", "--to-parent":
			d := DirParent
			dir = &d
		case "--down", "--to-children":
			d := DirChildren
			dir = &d
		case "--to-siblings":
			d := DirSiblings
			dir = &d
		case "--to-descendants":
			d := DirDescendants
			dir = &d
		case "--from":
			if i+1 < len(args) {
				from = args[i+1]
				i++
			}
		default:
			pos = append(pos, args[i])
		}
	}

	if dir != nil {
		cmdBroadcast(*dir, from, pos, mode)
		return
	}
	if len(pos) < 2 {
		fail("usage: csend send <cible> <message…>  (ou un drapeau famille: --up/--down/--to-siblings/--to-descendants)")
	}
	out, err := sendMessage(pos[0], strings.Join(pos[1:], " "), mode)
	if err != nil {
		fail(err.Error())
	}
	printOutcome("", out)
	if out.Action == "refused" {
		os.Exit(1)
	}
}

// cmdBroadcast delivers one message to a set of relatives (up/down the family).
func cmdBroadcast(dir Direction, from string, pos []string, mode SubmitMode) {
	if len(pos) < 1 {
		fail("usage: csend send --" + dir.String() + " <message…>")
	}
	text := strings.Join(pos, " ")
	baseKey := ""
	if from != "" {
		fs, err := resolveTarget(from)
		if err != nil {
			fail(err.Error())
		}
		baseKey = sessionKey(fs)
	} else {
		bk, err := selfKey()
		if err != nil {
			fail(err.Error())
		}
		baseKey = bk
	}
	targets, err := resolveFamily(baseKey, dir)
	if err != nil {
		fail(err.Error())
	}
	if len(targets) == 0 {
		fmt.Printf("Aucune cible en direction « %s » (lie des sessions avec `csend link`).\n", dir)
		return
	}
	fmt.Printf("Broadcast → %s (%d cible(s)):\n", dir, len(targets))
	failures := 0
	for _, t := range targets {
		out, err := deliverTo(t, text, mode)
		if err != nil {
			fmt.Printf("  ✗ %s (%s): %s\n", t.Ref, t.Workspace, err.Error())
			failures++
			continue
		}
		printOutcome("  ", out)
		if out.Action == "refused" {
			failures++
		}
	}
	if failures > 0 {
		os.Exit(1)
	}
}

func printOutcome(prefix string, out Outcome) {
	switch out.Action {
	case "submitted":
		fmt.Printf("%s✓ envoyé et validé → %s (%s) [%s]\n", prefix, out.Ref, out.Workspace, out.State)
	case "staged":
		fmt.Printf("%s• déposé (non validé) → %s (%s) [%s]%s\n", prefix, out.Ref, out.Workspace, out.State, reasonSuffix(out.Reason))
	case "refused":
		fmt.Printf("%s✗ refusé → %s (%s) [%s]%s\n", prefix, out.Ref, out.Workspace, out.State, reasonSuffix(out.Reason))
	}
}

func cmdRead(args []string) {
	lines := 30
	var pos []string
	for i := 0; i < len(args); i++ {
		if args[i] == "--lines" && i+1 < len(args) {
			fmt.Sscanf(args[i+1], "%d", &lines)
			i++
			continue
		}
		pos = append(pos, args[i])
	}
	if len(pos) < 1 {
		fail("usage: csend read <cible> [--lines N]")
	}
	tgt, err := resolveTarget(pos[0])
	if err != nil {
		fail(err.Error())
	}
	screen, err := readScreen(tgt.Ref, tgt.WorkspaceRef, lines)
	if err != nil {
		fail(err.Error())
	}
	fmt.Print(screen)
}

// --- small presentation helpers ---

func glyph(s State) string {
	switch s {
	case StateIdle:
		return "● idle"
	case StateBusy:
		return "◐ busy"
	case StateAwaitConfirm:
		return "⚠ confirm"
	case StateUnreachable:
		return "◌ unreach"
	default:
		return "· unknown"
	}
}

func shortID(id string) string {
	if id == "" {
		return "—"
	}
	if len(id) > 8 {
		return id[:8]
	}
	return id
}

// shortKey formats a family-graph node id: surface refs as-is, session-ids
// trimmed to a readable prefix.
func shortKey(id string) string {
	if strings.HasPrefix(id, "surface:") {
		return id
	}
	if len(id) > 14 {
		return id[:14]
	}
	return id
}

func truncate(s string, n int) string {
	r := []rune(s)
	if len(r) <= n {
		return s
	}
	return string(r[:n-1]) + "…"
}

func reasonSuffix(r string) string {
	if r == "" {
		return ""
	}
	return " — " + r
}

func fail(msg string) {
	fmt.Fprintln(os.Stderr, "csend: "+msg)
	os.Exit(1)
}
