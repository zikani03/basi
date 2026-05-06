package playwright

import (
	playwrightgo "github.com/playwright-community/playwright-go"
)

const catScript = `
() => {
    const landmarkTags = new Set(['NAV', 'MAIN', 'HEADER', 'FOOTER', 'SECTION', 'ARTICLE', 'FORM', 'ASIDE']);
    const interactiveTags = new Set(['BUTTON', 'A', 'INPUT', 'SELECT', 'TEXTAREA', 'DETAILS', 'SUMMARY']);
    // Priority attributes from SPEC.md
    const priorityAttrs = [
        'role', 'aria-label', 'aria-labelledby', 'name', 
        'type', 'title', 'placeholder', 'aria-selected', 'aria-expanded'
    ];

    function isInteresting(node) {
        if (node.nodeType !== Node.ELEMENT_NODE) return false;
        if (interactiveTags.has(node.tagName)) return true;
        if (landmarkTags.has(node.tagName)) return true;
        if (node.hasAttribute('role')) return true;
        if (node.hasAttribute('aria-label') || node.hasAttribute('aria-labelledby')) return true;
        if (node.tabIndex >= 0) return true;
        return false;
    }

    function serialize(node) {
        let tag = node.tagName.toLowerCase();
        let id = node.id ? '#' + node.id : '';
        let attrs = '';
        for (const attr of priorityAttrs) {
            if (node.hasAttribute(attr)) {
                let val = node.getAttribute(attr);
                // Escape quotes in attributes
                val = val.replace(/"/g, '&quot;');
                attrs += '[' + attr + '="' + val + '"]';
            }
        }
        return tag + id + attrs;
    }

    function walk(node) {
        if (node.nodeType !== Node.ELEMENT_NODE) return null;
        // Pruning: Skip non-visual or heavy non-interactive tags
        if (['SCRIPT', 'STYLE', 'NOSCRIPT', 'SVG', 'CANVAS', 'IFRAME'].includes(node.tagName)) return null;

        let childrenResults = [];
        for (let child of node.children) {
            let res = walk(child);
            if (res) childrenResults.push(res);
        }

        const selfInteresting = isInteresting(node);

        if (selfInteresting) {
            let serialized = serialize(node);
            if (childrenResults.length > 0) {
                // Use Emmet notation: parent>(child1+child2)
                const childrenStr = childrenResults.length > 1 
                    ? '(' + childrenResults.join('+') + ')' 
                    : childrenResults[0];
                return serialized + '>' + childrenStr;
            }
            return serialized;
        } else {
            // Pruning: If this node isn't interesting, flatten its children into the tree
            if (childrenResults.length === 0) return null;
            return childrenResults.join('+');
        }
    }

    return walk(document.body) || '';
}
`

// GenerateCAT generates a compressed accessibility tree in Emmet format for
// sending to an AI for auto-detecting a tag on a UI
func GenerateCAT(page playwrightgo.Page) (string, error) {
	result, err := page.Evaluate(catScript)
	if err != nil {
		return "", err
	}
	return result.(string), nil
}
