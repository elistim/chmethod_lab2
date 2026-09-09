'use strict';
const $ = id => document.getElementById(id);
const state = {tab: 'interpolation', interpolation: null, integration: null, defaults: null, busy: false, integrationRequested: false, table: null, charts: []};
const names = {simpson:'Симпсона',trapezoid:'Трапеций',midpoint:'Средних прямоугольников',gauss:'Гаусса — Лежандра'};
const colors = ['#2f6259','#e86f43','#3868b3'];
function displayExpression(source){return source.replace(/\^(-?\d+)/g,(_,power)=>[...power].map(c=>superscripts[c]).join('')).replaceAll('sqrt','√');}
function decimal(value){
  const text=String(value);if(!/[eE]/.test(text))return text;
  const [mantissa,exponent]=text.toLowerCase().split('e'),negative=mantissa.startsWith('-');
  const unsigned=negative?mantissa.slice(1):mantissa,parts=unsigned.split('.'),digits=parts.join(''),position=parts[0].length+Number(exponent);
  return (negative?'-':'')+(position<=0?'0.'+'0'.repeat(-position)+digits:position>=digits.length?digits+'0'.repeat(position-digits.length):digits.slice(0,position)+'.'+digits.slice(position));
}
const superscripts = {'0':'⁰','1':'¹','2':'²','3':'³','4':'⁴','5':'⁵','6':'⁶','7':'⁷','8':'⁸','9':'⁹','-':'⁻','+':'⁺'};
function format(x, digits=9) {
  if (x === null || x === undefined) return '—';
  if (typeof x !== 'number') return String(x);
  if (x === 0) return '0';
  if (Math.abs(x)<0.00001 || Math.abs(x)>=10000000) {
    const [mantissa, exponent] = x.toExponential(digits-1).split('e');
    const power=String(Number(exponent)).split('').map(c=>superscripts[c]).join('');
    return `${Number(mantissa)} × 10${power}`;
  }
  return Number(x.toPrecision(digits)).toString();
}
function status(message, kind='') { $('status').textContent=message; $('status').className='status '+kind; }
async function api(url, body) { const response=await fetch(url,body===undefined?{}:{method:'POST',headers:{'Content-Type':'application/json'},body:JSON.stringify(body)}); let data;try{data=await response.json();}catch{throw new Error('Сервер вернул некорректный ответ.');}if(!response.ok) throw new Error(data.error||'Ошибка сервера');return data; }
function number(id) {
  const text=$(id).value.trim();
  let normalized=text.replaceAll(',', '.').replaceAll('−','-').replace(/\s/g,'');
  normalized=normalized.replace(/[⁰¹²³⁴⁵⁶⁷⁸⁹⁻⁺]+/g, power=>'^'+[...power].map(c=>Object.keys(superscripts).find(key=>superscripts[key]===c)).join(''));
  const power=normalized.match(/^(?:(\d+(?:\.\d+)?)[×·*])?10\^([+-]?\d+)$/);
  const value=power ? Number(power[1]||1)*10**Number(power[2]) : Number(normalized);
  if(!text||!Number.isFinite(value))throw new Error('Заполните числовое поле: '+$(id).labels[0].textContent);
  return value;
}
function metric(label,value,note) {const card=document.createElement('article');card.className='card metric';for(const [cls,text] of [['metric-label',label],['metric-value',value],['metric-note',note]]){const el=document.createElement('div');el.className=cls;el.textContent=text;card.append(el);}return card;}
function summary(items) {$('summary').replaceChildren(...items.map(item=>metric(...item)));}
function table(id, headers, rows) {const root=$(id);root.replaceChildren();const head=document.createElement('thead'),tr=document.createElement('tr');for(const title of headers){const th=document.createElement('th');th.textContent=title;th.scope='col';tr.append(th);}head.append(tr);const body=document.createElement('tbody');for(const row of rows){const tr=document.createElement('tr');for(const value of row){const td=document.createElement('td');td.textContent=format(value,12);tr.append(td);}body.append(tr);}root.append(head,body);if(id==='results')state.table={headers,rows};}

// Self-contained SVG chart: no CDN and no frontend build step.
function chart(containerId, legendId, xs, series, options={}) {
  const container=$(containerId),legend=$(legendId);container.replaceChildren();legend.replaceChildren();
  const ns='http://www.w3.org/2000/svg',svg=document.createElementNS(ns,'svg');svg.setAttribute('role','img');svg.setAttribute('aria-label',options.label||'График');container.append(svg);
  const tip=document.createElement('div');tip.className='tooltip';tip.hidden=true;container.append(tip);
  const hidden=new Set();let currentGeometry;
  const node=(tag,attrs={},text)=>{const e=document.createElementNS(ns,tag);for(const [key,value]of Object.entries(attrs))e.setAttribute(key,value);if(text!==undefined)e.textContent=text;svg.append(e);return e;};
  function draw(){
    svg.replaceChildren();tip.hidden=true;const W=Math.max(container.clientWidth,250),H=container.clientHeight;svg.setAttribute('viewBox',`0 0 ${W} ${H}`);
    const pad={l:108,r:20,t:18,b:36},pw=W-pad.l-pad.r,ph=H-pad.t-pad.b;
    let xmin=Math.min(...xs),xmax=Math.max(...xs);if(xmin===xmax){xmin-=.5;xmax+=.5;}
    // Source nodes belong to P(x) and follow its visibility and scale.
    const visibleNodes=hidden.has(0)?[]:(options.nodes||[]);
    const ys=[...series.flatMap((s,i)=>hidden.has(i)?[]:s.values),...visibleNodes.map(p=>p.y)].filter(Number.isFinite);let ymin=ys.length?Math.min(...ys):0,ymax=ys.length?Math.max(...ys):1;
    if(options.zero)ymin=Math.min(0,ymin);const range=ymax-ymin||Math.max(Math.abs(ymax)*.1,1e-12);ymin-=range*.08;ymax+=range*.08;
    const px=x=>pad.l+(x-xmin)/(xmax-xmin)*pw,py=y=>pad.t+(ymax-y)/(ymax-ymin)*ph;
    for(let i=0;i<=4;i++){const y=ymin+(ymax-ymin)*i/4;node('line',{x1:pad.l,y1:py(y),x2:W-pad.r,y2:py(y),stroke:'#e1e6df'});node('text',{x:pad.l-10,y:py(y)+4,'text-anchor':'end'},format(y,5));}
    for(let i=0;i<=5;i++){const x=xmin+(xmax-xmin)*i/5;node('text',{x:px(x),y:H-13,'text-anchor':'middle'},format(x,5));}
    node('text',{x:W-pad.r,y:H-1,'text-anchor':'end',class:'axis-title'},options.xLabel||'x');
    const defs=node('defs'),clip=node('clipPath',{id:containerId+'-plot-clip',clipPathUnits:'userSpaceOnUse'});
    defs.append(clip);clip.append(node('rect',{x:pad.l-4,y:pad.t-4,width:pw+8,height:ph+8}));
    const plotLayer=node('g',{'clip-path':`url(#${containerId}-plot-clip)`});
    const plot=(tag,attrs)=>{const element=node(tag,attrs);plotLayer.append(element);return element;};
    series.forEach((s,j)=>{
      if(hidden.has(j))return;
      let path='';s.values.forEach((y,i)=>{if(!Number.isFinite(y))return;path+=(path?' L':'M')+px(xs[i]).toFixed(2)+','+py(y).toFixed(2);});
      if(options.area&&j===0)plot('path',{d:path+` L${px(xs[xs.length-1])},${py(0)} L${px(xs[0])},${py(0)} Z`,fill:'#2f62591c',stroke:'none'});
      plot('path',{d:path,fill:'none',stroke:s.color||colors[j%3],'stroke-width':2,'stroke-linejoin':'round',...(s.dash?{'stroke-dasharray':'6 4'}:{})});
      if(options.points)s.values.forEach((y,i)=>{if(Number.isFinite(y))plot('circle',{cx:px(xs[i]),cy:py(y),r:2.5,fill:s.color||colors[j%3]});});
    });
    for(const p of visibleNodes)plot('circle',{cx:px(p.x),cy:py(p.y),r:3.2,fill:'#fffefa',stroke:colors[0],'stroke-width':1.6});
    const cross=node('line',{x1:0,y1:pad.t,x2:0,y2:H-pad.b,stroke:'#899b94','stroke-dasharray':'3 3',visibility:'hidden'});currentGeometry={px,pad,pw,xmin,xmax,cross,W};
  }
  series.forEach((s,i)=>{const button=document.createElement('button');button.type='button';button.setAttribute('aria-pressed','true');const swatch=document.createElement('i');swatch.style.background=s.color||colors[i%3];button.append(swatch,document.createTextNode(s.name));button.onclick=()=>{if(hidden.has(i))hidden.delete(i);else hidden.add(i);button.classList.toggle('off',hidden.has(i));button.setAttribute('aria-pressed',String(!hidden.has(i)));draw();};legend.append(button);});
  svg.addEventListener('pointermove',event=>{const g=currentGeometry;if(!g)return;const rect=svg.getBoundingClientRect(),local=event.clientX-rect.left,target=g.xmin+(local-g.pad.l)/g.pw*(g.xmax-g.xmin);let index=0;for(let i=1;i<xs.length;i++)if(Math.abs(xs[i]-target)<Math.abs(xs[index]-target))index=i;g.cross.setAttribute('x1',g.px(xs[index]));g.cross.setAttribute('x2',g.px(xs[index]));g.cross.setAttribute('visibility','visible');tip.textContent=(options.xLabel||'x')+' = '+format(xs[index])+'\n'+series.filter((s,i)=>!hidden.has(i)).map(s=>s.name+': '+format(s.values[index],12)).join('\n');tip.hidden=false;tip.style.left=Math.max(0,Math.min(local+12,g.W-tip.offsetWidth-5))+'px';tip.style.top='8px';});
  svg.addEventListener('pointerleave',()=>{tip.hidden=true;currentGeometry?.cross.setAttribute('visibility','hidden');});
  const observer=new ResizeObserver(draw);observer.observe(container);state.charts.push(observer);draw();
}
function clearCharts(){state.charts.forEach(o=>o.disconnect());state.charts=[];}
function selectTab(tab){state.tab=tab;document.querySelectorAll('[data-tab]').forEach(b=>{const active=b.dataset.tab===tab;b.classList.toggle('active',active);b.setAttribute('aria-pressed',String(active));});$('workspace-caption').textContent={interpolation:'Приближение функций',differentiation:'Численное дифференцирование',integration:'Численное интегрирование'}[tab];const integration=tab==='integration';$('interpolation-form').hidden=integration;$('integration-form').hidden=!integration;$('derivative-view').hidden=tab!=='differentiation';render();if(integration&&!state.integration){state.integrationRequested=true;run('integration');}}
function render(){
  clearCharts();state.table=null;$('export').disabled=true;
  if(state.tab==='integration'){renderIntegration();return;}
  $('gauss-details').hidden=true;const data=state.interpolation;if(!data){status('Нажмите «Рассчитать», чтобы получить результаты.');return;}
  $('function-display').textContent='f(x) = '+displayExpression(data.expression);
  $('source-grid-display').textContent=`x ∈ [${format(data.nodes[0].x)}; ${format(data.nodes[data.nodes.length-1].x)}] · ${data.nodes.length} исходных узлов`;
  const derivative=state.tab==='differentiation',second=$('derivative-view').value==='second';
  const key=derivative?(second?'second':'first'):'value',exactKey=derivative?(second?'exactSecond':'exactFirst'):'exact',errorKey=derivative?(second?'errorSecond':'errorFirst'):'error';
  const label=derivative?(second?'P″(x)':'P′(x)'):'P(x)',exactLabel=derivative?(second?'f″(x)':'f′(x)'):'f(x)';
  summary(derivative?[
    ['Макс. ошибка f′',data.reference?format(data.maxError[1]):'—','На новой сетке'],['Макс. ошибка f″',data.reference?format(data.maxError[2]):'—','На новой сетке'],['Степень полинома',String(data.degree),'Производные полинома Ньютона']
  ]:[['Макс. погрешность',data.reference?format(data.maxError[0]):'—','max |P(xⱼ) − f(xⱼ)|'],['Новая сетка',String(data.rows.length),'Узлов для расчёта'],['Степень полинома',String(data.degree),`${data.degree+1} узлов в каждом полиноме`]]);
  $('chart-title').textContent=derivative?(second?'Вторая производная':'Первая производная'):'Приближение функции '+displayExpression(data.expression);if(!data.reference&&!derivative)$('chart-title').textContent='Приближение табличной функции';
  $('chart-subtitle').textContent=data.reference?'Полином Ньютона и аналитическая функция; кружки — исходные узлы':'Полином по пользовательской таблице; аналитическое сравнение отключено';if(derivative)$('chart-subtitle').textContent='Дифференцирование локальных полиномов; в местах смены узлов возможны скачки.';
  const series=[{name:label,values:data.curve.map(r=>r[key])}];if(data.reference)series.push({name:exactLabel,values:data.curve.map(r=>r[exactKey]),dash:true});
  chart('chart','chart-legend',data.curve.map(r=>r.x),series,{label:$('chart-title').textContent,nodes:derivative?null:data.nodes});
  $('error-card').hidden=!data.reference;
  if(data.reference){$('error-title').textContent='Абсолютная погрешность '+label;$('error-subtitle').textContent='Сравнение с аналитическими значениями на новой сетке';const sorted=[...data.rows].sort((a,b)=>a.x-b.x);chart('error-chart','error-legend',sorted.map(r=>r.x),[{name:`|${label} − ${exactLabel}|`,values:sorted.map(r=>r[errorKey]),color:colors[1]}],{points:true,zero:true});}
  $('table-title').textContent=derivative?'Первая и вторая производные':'Значения на новой сетке';$('table-subtitle').textContent=`Степень ${data.degree} · ${data.rows.length} строк · индексы исходных узлов с нуля`;
  if(derivative)table('results',['j','xⱼ','P′(xⱼ)','f′(xⱼ)','|Δf′|','P″(xⱼ)','f″(xⱼ)','|Δf″|','Узлы'],data.rows.map((r,i)=>[i,r.x,r.first,r.exactFirst,r.errorFirst,r.second,r.exactSecond,r.errorSecond,`${r.start}…${r.end}`]));
  else table('results',['j','xⱼ','f(xⱼ)','P(xⱼ)','|P − f|','Узлы'],data.rows.map((r,i)=>[i,r.x,r.exact,r.value,r.error,`${r.start}…${r.end}`]));
  $('explanation').innerHTML=derivative?'<p>Для каждой точки выбираются m + 1 ближайших исходных узлов. Строится полином Ньютона на разделённых разностях, затем вычисляются его первая и вторая производные. Дифференцируется сам полином, без разностного вычитания близких значений функции.</p><div class="math">f(x) = √x + 1<br>f′(x) = 1 / (2√x)<br>f″(x) = −1 / (4x√x)<br>Δ₁ = |P′ − f′|; Δ₂ = |P″ − f″|</div><p>Степень 1 даёт нулевую вторую производную полинома. Для второй производной используйте степень не ниже 2. Значения производных в исходных узлах не обязаны совпадать с точными.</p>':'<p>Используется «скользящий» полином Ньютона: для каждой точки новой сетки выбираются m + 1 ближайших узлов. При равных расстояниях выбирается левый узел. При степени 10 для варианта 9 получается один полином по всем 11 узлам.</p><div class="math">Pₘ(x) = c₀ + c₁(x−x₀) + … + cₘ∏ₖ₌₀ᵐ⁻¹(x−xₖ)<br>cₖ = f[x₀, …, xₖ]<br>Δ(xⱼ) = |Pₘ(xⱼ) − f(xⱼ)|</div><p>Исходная сетка: xᵢ = 1 + 0,1i, i = 0…10. Новая сетка: xⱼ = 1 + 0,05j, j = 0…20. Максимальная ошибка рассчитана только на новой сетке, а не на всём непрерывном отрезке. На равномерной сетке разделённые разности эквивалентны форме Ньютона с конечными разностями.</p>';
  if(!data.reference)$('explanation').insertAdjacentHTML('beforeend','<p>Для своей таблицы без известной аналитической функции точная погрешность не определяется; соответствующие поля обозначены «—».</p>');
  if(data.expression!=='sqrt(x)+1'){
    $('explanation').innerHTML='<p>По m + 1 ближайшим узлам строится полином Ньютона. Значения его первой и второй производных вычисляются аналитическим дифференцированием полинома.</p><div class="math">Pₘ(x) = c₀ + c₁(x−x₀) + … + cₘ∏ₖ₌₀ᵐ⁻¹(x−xₖ)<br>Δ = |Pₘ − f|; Δ₁ = |P′ₘ − f′|; Δ₂ = |P″ₘ − f″|</div><p>Производные введённой функции для сравнения вычисляются автоматическим дифференцированием выражения по правилам цепочки, произведения и частного. Это вычисления в арифметике float64. В точках недифференцируемости аналитическое сравнение недоступно.</p><p>Максимальные ошибки относятся к новой сетке. Если сравнение отключено, точные значения и погрешности обозначены «—». Сетки показаны в таблицах; после смены формулы исходные y пересчитываются, если включён соответствующий флажок.</p>';
  }
  $('export').disabled=false;status(`Расчёт выполнен: ${data.nodes.length} исходных узлов, ${data.rows.length} новых, степень ${data.degree}.`);
}
function renderIntegration(){
  const data=state.integration;$('gauss-details').hidden=true;
  if(!data){$('summary').replaceChildren();$('chart').replaceChildren();$('chart-legend').replaceChildren();$('error-card').hidden=true;$('results').replaceChildren();$('chart-title').textContent='Подынтегральная функция';$('chart-subtitle').textContent='Выберите метод и выполните расчёт';$('table-title').textContent='Уточнение шага';$('table-subtitle').textContent='';$('explanation').innerHTML='<p>Задание 6.5 включает две части: одну из формул Ньютона — Котеса с автоматическим выбором шага и формулу Гаусса. Переключите метод, чтобы выполнить обе части. Для варианта 9 начальное число интервалов и число узлов Гаусса установлены в 8.</p>';status('Выберите метод и нажмите «Вычислить интеграл».');return;}
  const gauss=data.input.method==='gauss';
  $('integrand-display').textContent='f(x) = '+displayExpression(data.input.expression);
  summary([['Значение интеграла',format(data.value,12),names[data.input.method]],[data.hasReference?'Относительная ошибка':(data.absoluteCriterion?'Оценка абсолютной ошибки':'Оценка относительной ошибки'),format(data.hasReference?data.relativeError:(data.absoluteCriterion?data.estimate:data.estimate/Math.abs(data.value))),`Требуется ≤ ${format(data.input.tolerance)}`],[gauss?'Конечное число панелей':'Конечное число интервалов',String(data.n),gauss?`${data.input.order} узлов на панель`:`h = ${format((data.input.b-data.input.a)/data.n)}`]]);
  $('chart-title').textContent='Подынтегральная функция';$('chart-subtitle').textContent=`f(x) = ${displayExpression(data.input.expression)} · [${format(data.input.a)}; ${format(data.input.b)}]`;
  chart('chart','chart-legend',data.curve.map(p=>p.x),[{name:'f(x)',values:data.curve.map(p=>p.y)}],{label:'Подынтегральная функция и площадь интеграла',area:true,zero:true});
  $('error-card').hidden=false;$('error-title').textContent='Сходимость интеграла';$('error-subtitle').textContent='При последовательном удвоении числа '+(gauss?'панелей':'интервалов');
  const convergence=[{name:'Iₙ',values:data.iterations.map(r=>r.value)}];if(data.hasReference)convergence.push({name:'Эталон',values:data.iterations.map(()=>data.reference),dash:true});
  chart('error-chart','error-legend',data.iterations.map((_,i)=>i),convergence,{points:true,xLabel:'итерация'});
  $('table-title').textContent='Журнал уточнения шага';$('table-subtitle').textContent=data.hasReference?`Эталон (степенной ряд): ${format(data.reference,15)} · |I − Iэт| = ${format(data.absoluteError)}`:'Для своей функции показана оценка метода; точная погрешность неизвестна. При |Iₙ| < 10⁻¹² критерий становится абсолютным.';
  table('results',['k',gauss?'Панели':'n','h','Iₙ',gauss?'|I₂ₙ − Iₙ|':'Оценка Рунге','Оценка критерия','Факт. отн. ошибка'],data.iterations.map((r,i)=>[i,r.n,r.h,r.value,r.estimate,r.relative,data.hasReference?Math.abs(r.value-data.reference)/Math.abs(data.reference):null]));
  const formula={simpson:'Iₙ = h/3 · [f(a) + f(b) + 4Σ f(x₂ᵢ₋₁) + 2Σ f(x₂ᵢ)]',trapezoid:'Iₙ = h · [(f(a) + f(b))/2 + Σᵢ₌₁ⁿ⁻¹ f(a+ih)]',midpoint:'Iₙ = h · Σᵢ₌₀ⁿ⁻¹ f(a + (i+½)h)',gauss:'I ≈ Σпанелей (h/2) · Σⱼ₌₁ᵐ wⱼ f(xсередина + h·tⱼ/2)'}[data.input.method];
  $('explanation').innerHTML=`<p>Формула ${names[data.input.method]}. ${gauss?'Узлы — корни полинома Лежандра; веса вычисляются по его производной. Первый расчёт — одна панель, то есть обычная формула Гаусса заданного порядка. Для уточнения панель делится пополам, порядок сохраняется.':'После начального расчёта число интервалов удваивается: n → 2n, шаг уменьшается вдвое.'}</p><div class="math">${formula}<br>${gauss?'R = |I₂ₙ − Iₙ|':'R = |I₂ₙ − Iₙ| / (2ᵖ − 1), p = '+(data.input.method==='simpson'?'4':'2')}<br>Критерий: R / |I₂ₙ| ≤ ε</div><p>${gauss?'Для Гаусса используется разность последовательных приближений без деления на 2²ᵐ − 1, чтобы не занижать оценку до наступления асимптотического режима.':'Оценка Рунге относится к уточнённому значению I₂ₙ; поправка Ричардсона к результату не добавляется.'} Это оценка погрешности, а не строгая граница. Дополнительно проверяется относительное отклонение от независимого эталона.</p><div class="math">Iэт = F(b) − F(a)<br>F(x) = Σₖ₌₀∞ C(2k,k) · x⁴ᵏ⁺¹ / [4ᵏ(4k+1)]</div><p>Эталон вычисляется сходящимся степенным рядом для |x| &lt; 1. Приложение допускает границы от −0,99 до 0,99. Максимум — 1 048 576 интервалов. Задание 6.5: выполните расчёт одной формулой Ньютона — Котеса, затем переключитесь на Гаусса с порядком 8.</p>`;
  if(gauss){$('gauss-details').hidden=false;table('gauss-table',['j','tⱼ','wⱼ'],data.gaussNodes.map((p,i)=>[i+1,p.x,p.y]));}
  if(!data.hasReference){
    $('explanation').innerHTML=`<p>Формула ${names[data.input.method]}. Число интервалов последовательно удваивается; для Гаусса удваивается число панелей при сохранении порядка.</p><div class="math">${formula}<br>${gauss?'R = |I₂ₙ − Iₙ|':'R = |I₂ₙ − Iₙ| / (2ᵖ − 1), p = '+(data.input.method==='simpson'?'4':'2')}<br>R / |I₂ₙ| ≤ ε; при |I₂ₙ| &lt; 10⁻¹² используется R ≤ ε.</div><p>Для введённой функции точный интеграл неизвестен. Показана оценка по последовательным приближениям, а не фактическая ошибка. Метод предназначен для непрерывных конечных функций на всём отрезке; разрывы и быстро осциллирующие функции могут сделать оценку ненадёжной. Максимум — 65 536 интервалов.</p>`;
  }
  $('export').disabled=false;status(data.message,data.converged?'':'error');
}
function interpolationInput(){
  const nodes=$('nodes').value.trim().split(/\r?\n/).filter(l=>l.trim()).map((line,i)=>{const tokens=line.trim().split(/[;\s]+/);if(tokens.length!==2||tokens.some(t=>!Number.isFinite(Number(t))))throw new Error(`Строка ${i+1}: нужны два числа x и y.`);return {x:Number(tokens[0]),y:Number(tokens[1])};});
  const degree=number('degree');if(!Number.isInteger(degree))throw new Error('Степень должна быть целой.');let grid=[];
  const custom=$('custom-grid').value.trim();if(custom){grid=custom.split(/[;\s]+/).map(Number);if(grid.some(x=>!Number.isFinite(x)))throw new Error('В новой сетке есть некорректное число.');}
  else{const a=number('grid-a'),b=number('grid-b'),h=number('grid-step');if(b<a||h<=0)throw new Error('Проверьте границы и положительный шаг новой сетки.');const count=Math.round((b-a)/h);if(count>2000||Math.abs(a+count*h-b)>1e-9*Math.max(1,Math.abs(b)))throw new Error('Шаг должен делить отрезок на целое число интервалов (не более 2000).');grid=Array.from({length:count+1},(_,i)=>i===count?b:a+i*h);}
  return {nodes,grid,degree,reference:$('reference').checked,expression:$('function-expression').value.trim(),useFunction:$('use-function').checked};
}
async function run(kind){
  if(state.busy)return;if(kind==='integration')state.integrationRequested=false;state.busy=true;document.querySelectorAll('form button').forEach(b=>b.disabled=true);status('Выполняется расчёт…','pending');
  try{if(kind==='integration'){state.integration=await api('/api/integrate',{a:number('int-a'),b:number('int-b'),n:number('intervals'),order:number('order'),tolerance:number('tolerance'),method:$('method').value,expression:$('integrand-expression').value.trim()});}else{const input=interpolationInput();state.interpolation=await api('/api/interpolate',input);if(input.useFunction)$('nodes').value=state.interpolation.nodes.map(p=>`${decimal(p.x)}; ${decimal(p.y)}`).join('\n');}render();}
  catch(error){status(error.message+' Предыдущие результаты, если есть, сохранены.','error');}
  finally{state.busy=false;document.querySelectorAll('form button').forEach(b=>b.disabled=false);if(state.integrationRequested){state.integrationRequested=false;if(!state.integration)run('integration');}}
}
function resetInterpolation(){const d=state.defaults;if(!d)return;$('function-expression').value=d.expression;$('use-function').checked=true;$('degree').value=d.degree;$('degree').max=10;$('nodes').value=d.nodes.map(p=>`${decimal(p.x)}; ${decimal(p.y)}`).join('\n');$('grid-a').value=1;$('grid-b').value=2;$('grid-step').value=.05;$('custom-grid').value='';$('reference').checked=true;}
function methodChanged(){const gauss=$('method').value==='gauss';$('order-field').hidden=!gauss;$('n-field').hidden=gauss;markDirty();}
function markDirty(){if(!state.busy)status('Параметры изменены. Нажмите кнопку расчёта; ниже показан предыдущий результат.','pending');}
document.querySelectorAll('[data-tab]').forEach(button=>button.onclick=()=>selectTab(button.dataset.tab));
$('interpolation-form').onsubmit=event=>{event.preventDefault();run('interpolation');};$('integration-form').onsubmit=event=>{event.preventDefault();run('integration');};
$('reset').onclick=()=>{resetInterpolation();run('interpolation');};$('reset-integration').onclick=()=>{$('integration-form').reset();methodChanged();run('integration');};
$('nodes').addEventListener('input',()=>{$('degree').max=Math.min(20,$('nodes').value.trim().split(/\r?\n/).filter(l=>l.trim()).length-1);});
document.querySelectorAll('form').forEach(form=>form.addEventListener('input',markDirty));$('method').onchange=methodChanged;$('derivative-view').onchange=render;
$('export').onclick=()=>{if(!state.table)return;const escape=value=>'"'+(typeof value==='number'?decimal(value):String(value??'')).replaceAll('"','""')+'"';const metadata=state.tab==='integration'?[['Задание','6.5'],['Метод',names[state.integration.input.method]],['Функция',state.integration.input.expression],['a',state.integration.input.a],['b',state.integration.input.b],['epsilon',state.integration.input.tolerance],['Порядок Гаусса',state.integration.input.order]]:[['Задание',state.tab==='interpolation'?'6.3':'6.4'],['Степень',state.interpolation.degree],['Функция',state.interpolation.expression],['Аналитическое сравнение',state.interpolation.reference],['Исходная таблица'],['x','y'],...state.interpolation.nodes.map(p=>[p.x,p.y])];const csv='\uFEFF'+[...metadata,[],state.table.headers,...state.table.rows].map(row=>row.map(escape).join(';')).join('\r\n');const url=URL.createObjectURL(new Blob([csv],{type:'text/csv;charset=utf-8'}));const link=document.createElement('a');link.href=url;link.download=`variant9-${state.tab}.csv`;link.click();setTimeout(()=>URL.revokeObjectURL(url),1000);};
$('print').onclick=()=>window.print();
async function init(){try{state.defaults=await api('/api/defaults');resetInterpolation();await run('interpolation');}catch(error){status('Не удалось загрузить данные: '+error.message,'error');}}
init();
