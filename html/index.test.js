const test = require('node:test');
const assert = require('node:assert/strict');
const fs = require('node:fs');
const path = require('node:path');
const vm = require('node:vm');

function extractFunction(source, name) {
    const marker = `function ${name}(`;
    const start = source.indexOf(marker);

    if (start === -1) {
        return null;
    }

    let depth = 0;
    let end = start;
    let bodyStarted = false;

    for (let i = start; i < source.length; i += 1) {
        const char = source[i];

        if (char === '{') {
            depth += 1;
            bodyStarted = true;
        } else if (char === '}') {
            depth -= 1;

            if (bodyStarted && depth === 0) {
                end = i + 1;
                break;
            }
        }
    }

    return source.slice(start, end);
}

function loadParserFunctions() {
    const html = fs.readFileSync(path.join(__dirname, 'index.html'), 'utf8');
    const script = html.match(/<script>([\s\S]*)<\/script>/)?.[1];

    if (!script) {
        throw new Error('Unable to locate inline script in html/index.html');
    }

    const functionNames = [
        'getDirectChildByTagName',
        'getChannelImageUrl',
        'parseRSSEpisodes'
    ];
    const extracted = functionNames
        .map(name => extractFunction(script, name))
        .filter(Boolean)
        .join('\n\n');

    const context = { console };
    vm.runInNewContext(`${extracted}\nthis.parseRSSEpisodes = parseRSSEpisodes;`, context);
    return context.parseRSSEpisodes;
}

function makeTextNode(textContent) {
    return { textContent };
}

function makeImageNode(href) {
    return {
        getAttribute(name) {
            return name === 'href' ? href : null;
        }
    };
}

function makeItem(options = {}) {
    const values = {
        title: options.title || 'Episode',
        description: options.description || '',
        link: options.link || '',
        pubDate: options.pubDate || '',
        guid: options.guid || '',
        author: options.author || '',
        'itunes\\:author, author[itunes]': options.itunesAuthor || '',
        'itunes\\:duration, duration[itunes]': options.duration || ''
    };

    return {
        querySelector(selector) {
            if (selector === 'enclosure') {
                return null;
            }

            if (selector in values && values[selector] !== '') {
                return makeTextNode(values[selector]);
            }

            return null;
        },
        getElementsByTagName(tagName) {
            if (tagName === 'itunes:image' && options.imageHref) {
                return [makeImageNode(options.imageHref)];
            }

            return [];
        }
    };
}

function makeChannel({ items, descendantItunesImageHref, rssImageUrl, directItunesImageHref }) {
    return {
        children: [
            ...(directItunesImageHref ? [{ tagName: 'itunes:image', getAttribute: name => name === 'href' ? directItunesImageHref : null }] : []),
            ...(rssImageUrl ? [{ tagName: 'image', children: [{ tagName: 'url', textContent: rssImageUrl }] }] : []),
            ...items.map(item => ({ tagName: 'item', __item: item }))
        ],
        querySelector(selector) {
            const values = {
                title: 'Feed Title',
                description: 'Feed Description',
                link: 'https://example.com/feed'
            };

            return selector in values ? makeTextNode(values[selector]) : null;
        },
        querySelectorAll(selector) {
            return selector === 'item' ? items : [];
        },
        getElementsByTagName(tagName) {
            if (tagName === 'itunes:image' && descendantItunesImageHref) {
                return [makeImageNode(descendantItunesImageHref)];
            }

            return [];
        }
    };
}

function makeRssDoc(channel) {
    return {
        querySelector(selector) {
            return selector === 'channel' ? channel : null;
        }
    };
}

test('parseRSSEpisodes falls back to the channel RSS image url', () => {
    const parseRSSEpisodes = loadParserFunctions();
    const item = makeItem();
    const channel = makeChannel({
        items: [item],
        rssImageUrl: 'https://example.com/channel-rss.jpg'
    });

    const episodes = parseRSSEpisodes(makeRssDoc(channel), 'https://example.com/feed.xml');

    assert.equal(episodes[0].image, 'https://example.com/channel-rss.jpg');
});

test('parseRSSEpisodes ignores descendant item itunes images for channel fallback', () => {
    const parseRSSEpisodes = loadParserFunctions();
    const firstItem = makeItem({ title: 'First Episode' });
    const secondItem = makeItem({ title: 'Second Episode', imageHref: 'https://example.com/second-episode.jpg' });
    const channel = makeChannel({
        items: [firstItem, secondItem],
        descendantItunesImageHref: 'https://example.com/second-episode.jpg'
    });

    const episodes = parseRSSEpisodes(makeRssDoc(channel), 'https://example.com/feed.xml');

    assert.equal(episodes[0].image, '');
    assert.equal(episodes[1].image, 'https://example.com/second-episode.jpg');
});
