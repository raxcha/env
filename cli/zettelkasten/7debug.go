package zettelkasten

import "env/filesystem"

const zettelDebugData = true

func zettelDebugRandPage() *filesystem.Page {
	note := func(name, path string, tags []string) *filesystem.Page {
		return &filesystem.Page{
			Name:     name,
			Path:     "c.rand/" + path,
			Type:     "note",
			Metadata: map[string]any{"tags": tags},
		}
	}

	return &filesystem.Page{
		Name: "c.rand",
		Path: "c.rand",
		Type: "deep",
		Children: []*filesystem.Page{
			note("The Turing Test", "turing-test", []string{"ai", "philosophy", "computing"}),
			note("Gödel's Incompleteness", "godel", []string{"mathematics", "logic", "philosophy"}),
			note("Lambda Calculus", "lambda-calculus", []string{"computing", "mathematics", "functional"}),
			note("The Selfish Gene", "selfish-gene", []string{"biology", "evolution", "science"}),
			note("Meditations", "meditations", []string{"philosophy", "stoicism", "classics"}),
			note("Phenomenology of Spirit", "hegel-phenomenology", []string{"philosophy", "idealism", "history"}),
			note("On the Origin of Species", "darwin-origin", []string{"biology", "evolution", "science", "history"}),
			note("The Structure of Scientific Revolutions", "kuhn", []string{"science", "philosophy", "history"}),
			note("Critique of Pure Reason", "kant-critique", []string{"philosophy", "epistemology", "classics"}),
			note("Principia Mathematica", "principia", []string{"mathematics", "logic", "classics"}),
			note("Being and Time", "heidegger", []string{"philosophy", "phenomenology", "existentialism"}),
			note("The Republic", "plato-republic", []string{"philosophy", "politics", "classics"}),
			note("Nicomachean Ethics", "aristotle-ethics", []string{"philosophy", "ethics", "classics"}),
			note("The Wealth of Nations", "smith-wealth", []string{"economics", "history", "politics"}),
			note("Capital", "marx-capital", []string{"economics", "politics", "history"}),
			note("The General Theory", "keynes-general", []string{"economics", "politics"}),
			note("On Liberty", "mill-liberty", []string{"philosophy", "politics", "ethics"}),
			note("Leviathan", "hobbes-leviathan", []string{"philosophy", "politics", "classics"}),
			note("Social Contract", "rousseau-social-contract", []string{"philosophy", "politics"}),
			note("Thus Spoke Zarathustra", "nietzsche-zarathustra", []string{"philosophy", "existentialism", "literature"}),
			note("The Iliad", "iliad", []string{"literature", "poetry", "classics"}),
			note("Divine Comedy", "dante-comedy", []string{"literature", "poetry", "classics"}),
			note("Hamlet", "shakespeare-hamlet", []string{"literature", "drama", "classics"}),
			note("Don Quixote", "cervantes", []string{"literature", "fiction", "classics"}),
			note("Ulysses", "joyce-ulysses", []string{"literature", "fiction", "modernism"}),
			note("In Search of Lost Time", "proust", []string{"literature", "fiction", "modernism"}),
			note("The Brothers Karamazov", "dostoevsky", []string{"literature", "fiction", "existentialism"}),
			note("War and Peace", "tolstoy", []string{"literature", "fiction", "history"}),
			note("One Hundred Years of Solitude", "garcia-marquez", []string{"literature", "fiction", "magical-realism"}),
			note("The Name of the Rose", "eco-rose", []string{"literature", "fiction", "history"}),
			note("A General Theory of Love", "general-love", []string{"psychology", "neuroscience", "relationships"}),
			note("The Interpretation of Dreams", "freud-dreams", []string{"psychology", "psychoanalysis"}),
			note("Being Mortal", "gawande-mortal", []string{"medicine", "ethics", "psychology"}),
			note("The Man Who Mistook His Wife for a Hat", "sacks", []string{"neuroscience", "psychology", "medicine"}),
			note("Thinking Fast and Slow", "kahneman", []string{"psychology", "economics", "science"}),
			note("The Structure of Magic", "bandler", []string{"psychology", "linguistics", "communication"}),
			note("Gödel, Escher, Bach", "hofstadter", []string{"computing", "mathematics", "philosophy", "music"}),
			note("The Art of Computer Programming", "knuth", []string{"computing", "mathematics", "algorithms"}),
			note("Structure and Interpretation of Computer Programs", "sicp", []string{"computing", "functional", "mathematics"}),
			note("Clean Code", "clean-code", []string{"computing", "software", "engineering"}),
			note("The Pragmatic Programmer", "pragmatic", []string{"computing", "software", "engineering"}),
			note("Design Patterns", "gang-of-four", []string{"computing", "software", "engineering"}),
			note("A Brief History of Time", "hawking", []string{"physics", "science", "cosmology"}),
			note("The Elegant Universe", "elegant-universe", []string{"physics", "science", "cosmology"}),
			note("The Feynman Lectures", "feynman", []string{"physics", "science", "education"}),
			note("The Double Helix", "double-helix", []string{"biology", "science", "history"}),
			note("What Is Life?", "schrodinger-life", []string{"biology", "physics", "science"}),
			note("The Well-Tempered Clavier", "bach-wtc", []string{"music", "baroque", "keyboard"}),
			note("Symphony No. 9", "beethoven-9", []string{"music", "classical", "orchestral"}),
			note("Kind of Blue", "miles-kind-of-blue", []string{"music", "jazz", "improvisation"}),
			note("Revolver", "beatles-revolver", []string{"music", "rock", "studio"}),
			note("Disintegration", "cure-disintegration", []string{"music", "rock", "gothic"}),
			note("OK Computer", "radiohead-ok", []string{"music", "rock", "electronic"}),
			note("Guernica", "picasso-guernica", []string{"art", "painting", "politics"}),
			note("The Starry Night", "van-gogh-starry", []string{"art", "painting", "impressionism"}),
			note("The Persistence of Memory", "dali", []string{"art", "painting", "surrealism"}),
			note("Fountain", "duchamp-fountain", []string{"art", "conceptual", "modernism"}),
			note("Blade Runner", "blade-runner", []string{"film", "sci-fi", "ai"}),
			note("2001: A Space Odyssey", "kubrick-2001", []string{"film", "sci-fi", "ai"}),
			note("Stalker", "tarkovsky-stalker", []string{"film", "sci-fi", "philosophy"}),
			note("Rashomon", "kurosawa-rashomon", []string{"film", "drama", "philosophy"}),
			note("Persona", "bergman-persona", []string{"film", "drama", "psychology"}),
		},
	}
}

func zettelDebugFamiPage() *filesystem.Page {
	family := func(name, path string, children ...string) *filesystem.Page {
		page := &filesystem.Page{
			Name:     name,
			Path:     "d.fami/" + path,
			Type:     "note",
			Metadata: map[string]any{"name": name},
		}
		for _, childName := range children {
			page.Children = append(page.Children, &filesystem.Page{
				Name:     childName,
				Path:     "d.fami/" + path + "/" + childName,
				Type:     "note",
				Metadata: map[string]any{"name": childName},
			})
		}
		return page
	}

	return &filesystem.Page{
		Name: "d.fami",
		Path: "d.fami",
		Type: "deep",
		Children: []*filesystem.Page{
			family("Ancient Philosophy", "ancient-philosophy",
				"Pre-Socratics", "Socrates", "Plato", "Aristotle", "Stoics", "Epicureans", "Skeptics"),
			family("Modern Philosophy", "modern-philosophy",
				"Descartes", "Spinoza", "Leibniz", "Locke", "Hume", "Kant", "Hegel"),
			family("Continental Philosophy", "continental",
				"Nietzsche", "Heidegger", "Sartre", "Merleau-Ponty", "Derrida", "Deleuze", "Foucault"),
			family("Analytic Philosophy", "analytic",
				"Frege", "Russell", "Wittgenstein", "Carnap", "Quine", "Davidson", "Kripke"),
			family("Foundations of Mathematics", "foundations-math",
				"Set Theory", "Category Theory", "Type Theory", "Proof Theory", "Model Theory"),
			family("Programming Paradigms", "programming-paradigms",
				"Imperative", "Object-Oriented", "Functional", "Logic", "Concurrent", "Reactive"),
			family("Machine Learning", "machine-learning",
				"Supervised", "Unsupervised", "Reinforcement", "Deep Learning", "Transformers", "Diffusion"),
			family("Distributed Systems", "distributed-systems",
				"Consensus", "CAP Theorem", "Event Sourcing", "CRDTs", "Service Mesh", "Observability"),
			family("Evolutionary Biology", "evolutionary-biology",
				"Natural Selection", "Genetic Drift", "Speciation", "Evo-Devo", "Phylogenetics", "Coevolution"),
			family("Neuroscience", "neuroscience",
				"Neurons", "Synaptic Plasticity", "Memory Consolidation", "Attention", "Consciousness", "Sleep"),
			family("Quantum Mechanics", "quantum",
				"Wave Functions", "Superposition", "Entanglement", "Measurement Problem", "Decoherence"),
			family("Cosmology", "cosmology",
				"Big Bang", "Inflation", "Dark Matter", "Dark Energy", "Black Holes", "Multiverse"),
			family("Classical Music", "classical-music",
				"Baroque", "Classical Period", "Romanticism", "Modernism", "Minimalism", "Contemporary"),
			family("Jazz History", "jazz-history",
				"Blues Roots", "Swing Era", "Bebop", "Modal Jazz", "Free Jazz", "Fusion", "Contemporary"),
			family("Cinema Movements", "cinema",
				"Soviet Montage", "Italian Neorealism", "French New Wave", "New Hollywood", "Dogme 95"),
			family("Political Philosophy", "political-philosophy",
				"Liberalism", "Conservatism", "Socialism", "Anarchism", "Communitarianism", "Republicanism"),
			family("Ethics", "ethics",
				"Consequentialism", "Deontology", "Virtue Ethics", "Contractualism", "Care Ethics"),
			family("Linguistics", "linguistics",
				"Phonology", "Morphology", "Syntax", "Semantics", "Pragmatics", "Sociolinguistics"),
			family("Economic Schools", "economics",
				"Classical", "Keynesian", "Monetarism", "Austrian", "Institutional", "Behavioral"),
			family("Cognitive Science", "cognitive-science",
				"Perception", "Attention", "Memory", "Language", "Reasoning", "Decision Making"),
			family("Graph Theory", "graph-theory",
				"Trees", "Flows", "Matchings", "Planarity", "Coloring", "Spectral Theory"),
			family("Topology", "topology",
				"Point-Set", "Algebraic", "Differential", "Geometric", "K-Theory"),
			family("Complexity Theory", "complexity",
				"P vs NP", "NP-Complete", "Approximation", "Parameterized", "Communication Complexity"),
			family("Philosophy of Mind", "philosophy-of-mind",
				"Dualism", "Physicalism", "Functionalism", "Phenomenal Consciousness", "Intentionality"),
		},
	}
}
