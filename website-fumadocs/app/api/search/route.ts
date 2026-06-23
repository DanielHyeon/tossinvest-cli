import { source } from '@/lib/source';
import { createFromSource } from 'fumadocs-core/search/server';

export const revalidate = false;

export const { staticGET: GET } = createFromSource(source, {
  // Orama has no Korean tokenizer; map both locales to a supported language.
  // https://docs.orama.com/docs/orama-js/supported-languages
  localeMap: {
    ko: 'english',
    en: 'english',
  },
});
