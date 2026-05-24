// ПИВО
const BEERS = [
    { id:1, name:'Karmeliet Tripel', country:'Бельгия', abv:'8.4', ibu:22, emoji:'🍺', style:'IPA', desc:'Пряный трапистский эль на трёх злаках' },
    { id:2, name:'Duvel Single', country:'Бельгия', abv:'6.8', ibu:30, emoji:'🍻', style:'IPA', desc:'Светлый бельгийский эль с цветочными нотками хмеля' },
    { id:3, name:'Leffe Blonde', country:'Бельгия', abv:'6.6', ibu:20, emoji:'🍺', style:'IPA', desc:'Мягкий и сладковатый аббатский эль' },
    { id:4, name:'Westmalle Tripel', country:'Бельгия', abv:'9.5', ibu:36, emoji:'🍺', style:'IPA', desc:'Классический цистерцианский трапист' },
    { id:5, name:'Chimay Blue', country:'Бельгия', abv:'9.0', ibu:30, emoji:'🍻', style:'IPA', desc:'Тёмный трапистский эль' },
    { id:6, name:'Orval', country:'Бельгия', abv:'6.2', ibu:28, emoji:'🍺', style:'IPA', desc:'Уникальный трапист с диким дрожжевым характером' },
    { id:7, name:'Rochefort 10', country:'Бельгия', abv:'11.3', ibu:27, emoji:'🍻', style:'IPA', desc:'Один из самых крепких трапистов' },
];

// ОТЗЫВЫ (ключ = id пива)
const REVIEWS_DATA = {
    1: [
        { author:'Алексей К.', stars:5, text:'Отличное бельгийское, мягкий вкус с пряными нотами.' },
        { author:'Мария П.', stars:4, text:'Чуть сладковатое, но в целом достойное.' }
    ],
    2: [{ author:'Игорь С.', stars:5, text:'Мой любимый Duvel — освежающий с хорошей горчинкой.' }],
};

// ПАРАМЕТРЫ КАТЕГОРИЙ
let PARAMS = [
    { field:'abv', type:'numeric', range:'0 — 20', inherited:true },
    { field:'ibu', type:'numeric', range:'0 — 120', inherited:true },
    { field:'country', type:'enum', range:'CountryEnum', inherited:false },
    { field:'style', type:'enum', range:'BeerStyleEnum', inherited:true },
    { field:'og', type:'numeric', range:'1.000 — 1.200', inherited:false },
];

// ENUM-СПРАВОЧНИКИ
const ENUM_DATA = {
    'CountryEnum': ['Бельгия','Германия','Чехия','Россия','США','Великобритания','Нидерланды','Австрия','Япония','Ирландия','Дания','Польша'],
    'BeerStyleEnum': ['IPA','Пейл эль','Сэзон','Трапист','Дуббель','Триппель','Квадрупель','Гёз'],
    'ColorEnum': ['Светлое','Золотистое','Янтарное','Медное','Тёмное','Чёрное'],
};

let activeEnum = 'CountryEnum';

// СОХРАНЕНИЕ/ЗАГРУЗКА В LOCALSTORAGE
function saveParams() { localStorage.setItem('beer_params', JSON.stringify(PARAMS)); }
function loadParams() {
    const saved = localStorage.getItem('beer_params');
    if (saved) PARAMS = JSON.parse(saved);
}
loadParams();

function saveEnums() { localStorage.setItem('beer_enums', JSON.stringify(ENUM_DATA)); }
function loadEnums() {
    const saved = localStorage.getItem('beer_enums');
    if (saved) {
        const loaded = JSON.parse(saved);
        Object.keys(loaded).forEach(k => { ENUM_DATA[k] = loaded[k]; });
    }
}
loadEnums();

function saveReviews() { localStorage.setItem('beer_reviews', JSON.stringify(REVIEWS_DATA)); }
function loadReviews() {
    const saved = localStorage.getItem('beer_reviews');
    if (saved) {
        const loaded = JSON.parse(saved);
        Object.keys(loaded).forEach(k => { REVIEWS_DATA[k] = loaded[k]; });
    }
}
loadReviews();