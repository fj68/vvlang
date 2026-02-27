import { getHighlighter } from 'https://esm.sh/shiki@1.0.0';

async function initShiki() {
    // 1. Fetch the language definition
    const langUrl = new URL('vv.tmLanguage.json', import.meta.url).href;
    const response = await fetch(langUrl);
    const vvLang = await response.json();
    
    // 2. Initialize Shiki
    const highlighter = await getHighlighter({
        themes: ['github-light', 'github-dark'],
        langs: [vvLang]
    });

    // 3. Find and replace code blocks
    const preElements = document.querySelectorAll('pre');
    preElements.forEach((pre) => {
        const codeElement = pre.querySelector('code');
        if (!codeElement) return;

        const codeContent = codeElement.textContent || codeElement.innerText;
        
        // Add a wrapper for relative positioning of the copy button
        const wrapper = document.createElement('div');
        wrapper.className = 'shiki-wrapper';
        
        // Create the copy button
        const copyBtn = document.createElement('button');
        copyBtn.className = 'shiki-copy-btn';
        copyBtn.innerText = 'Copy';
        copyBtn.title = 'Copy code';
        
        copyBtn.addEventListener('click', async () => {
            try {
                await navigator.clipboard.writeText(codeContent);
                copyBtn.innerText = 'Copied!';
                setTimeout(() => { copyBtn.innerText = 'Copy'; }, 2000);
            } catch (err) {
                console.error('Failed to copy text: ', err);
            }
        });

        // 4. Generate HTML from Shiki
        // We use string manipulation to trim any trailing newlines from the code before highlighting,
        // otherwise Shiki might render an extra empty line.
        const trimmedCode = codeContent.replace(/\n$/, '');

        const html = highlighter.codeToHtml(trimmedCode, { 
            lang: 'vv',
            themes: {
                light: 'github-light',
                dark: 'github-dark'
            }
        });
        
        const div = document.createElement('div');
        div.innerHTML = html;
        const shikiPre = div.firstElementChild;
        
        // Move any custom styles/classes from original pre if needed, 
        // but generally shiki handles the styling.
        
        wrapper.appendChild(copyBtn);
        wrapper.appendChild(shikiPre);
        
        // 5. Replace original pre with the Shiki wrapper
        pre.parentNode.replaceChild(wrapper, pre);
    });
}

// Support for dark mode switching if the user's OS prefers dark mode
// Shiki 'themes' config supports multi-theme out of the box in newer versions with css variables,
// but for v1.0.0 we might need to rely on the standard CSS query if configured via css-variables,
// or we just inject github-light/dark and CSS will handle dark mode via media queries if we config it.
// According to Shiki docs, to use `themes` mapping we just need to ensure our CSS supports the data-theme attribute,
// or use `css-variables` theme. The `themes` block above with `light` and `dark` options creates HTML with
// inline styles `style="--shiki-light:...; --shiki-dark:..."`.

document.addEventListener('DOMContentLoaded', initShiki);
