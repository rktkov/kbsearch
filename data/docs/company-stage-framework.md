# Company Stage Framework — Контекст инженерных решений

**Статус:** Draft 1 — на ревью  
**Дата:** 2026-09-06  
**Продукт:** I Am Lead  
**Связанные документы:** `product-overview.md`, `learning-architecture.md`

---

## 1. Назначение документа

Этот документ задаёт контекст, в котором участник должен применять инженерные практики.

Одна и та же практика — PRD, RFC, тестовая стратегия, rollout или postmortem — не имеет универсальной глубины. Её форма зависит от:

- стоимости ошибки;
- масштаба системы;
- количества команд и зависимостей;
- требований к скорости принятия решений;
- регуляторных ограничений;
- стоимости простоя;
- зрелости процессов;
- доступного инженерного ресурса.

Главный принцип:

> Зрелость — это не максимальная формальность. Зрелость — это способность выбрать достаточный уровень формальности для конкретного контекста.

---

## 2. Как использовать framework

Перед созданием документа или проектированием решения участник отвечает на пять вопросов:

1. На какой стадии находится компания или конкретный бизнес-направление?
2. Какова цена ошибки и цена задержки?
3. Сколько команд и систем затрагивает решение?
4. Какие обязательные ограничения есть у бизнеса, безопасности и регуляторов?
5. Какой минимальный уровень документации позволит принять, реализовать и сопровождать решение?

Framework не заменяет инженерное суждение. Он помогает объяснить, почему выбранная глубина является разумной.

---

## 3. Стадии компаний

### 3.1. Pre-seed / Seed

#### Контекст

- гипотеза продукта ещё проверяется;
- команда небольшая;
- многие роли совмещены;
- архитектура быстро меняется;
- стоимость задержки часто выше стоимости локального технического долга.

#### Главный инженерный фокус

- скорость проверки гипотезы;
- минимальный достаточный scope;
- сохранение возможности быстро изменить направление;
- отсутствие необратимых решений без необходимости.

#### Что обычно не нужно

- длинные approval chains;
- сложные комитеты;
- документация ради документации;
- enterprise-grade процесс для некритичного MVP.

#### Риск чрезмерной формальности

Команда тратит больше времени на согласование, чем на проверку продукта.

---

### 3.2. Series A

#### Контекст

- product-market fit начинает подтверждаться;
- растёт команда;
- появляются повторяемые процессы;
- увеличивается стоимость ошибок;
- отдельные компоненты начинают становиться критичными.

#### Главный инженерный фокус

- повторяемая delivery-модель;
- снижение rework;
- понятные ownership и зависимости;
- базовая наблюдаемость и эксплуатация.

#### Что начинает становиться обязательным

- PRD с явным scope;
- RFC для значимых технических решений;
- ADR для решений, которые нужно помнить через полгода;
- Test Strategy для критических функций;
- понятный rollout и rollback.

#### Риск недостаточной зрелости

Команда продолжает работать как маленький стартап, хотя количество зависимостей уже выросло.

---

### 3.3. Scale-up

#### Контекст

- несколько продуктовых и инженерных команд;
- параллельная разработка;
- большое количество интеграций;
- несколько production-критичных систем;
- решения одной команды влияют на другие команды.

#### Главный инженерный фокус

- независимость команд;
- управление межкомандными зависимостями;
- стабильные API и event contracts;
- владение системами и данными;
- предсказуемый rollout.

#### Что становится обязательным

- Design Docs и RFC;
- ADR trail;
- Context Maps;
- RACI и dependency maps;
- SLO/SLI для критичных систем;
- migration и compatibility plans;
- регулярные postmortems.

#### Риск недостаточной зрелости

Команды оптимизируют локальные решения, создавая системные зависимости и coordination bottlenecks.

---

### 3.4. Corporation / Enterprise

#### Контекст

- много команд и организационных уровней;
- длительный жизненный цикл систем;
- compliance, legal и security requirements;
- высокая цена breaking changes;
- необходимость audit trail и формального ownership.

#### Главный инженерный фокус

- управляемость изменений;
- совместимость;
- безопасность и соответствие требованиям;
- прозрачность решений;
- снижение организационного и операционного риска.

#### Что обычно требуется

- формальные RFC review;
- architecture governance;
- security и privacy review;
- ADR и decision history;
- API/versioning policy;
- formal release gates;
- disaster recovery evidence;
- incident communication protocol.

#### Риск чрезмерной зрелости

Governance становится самоцелью и блокирует изменения, не снижая реальных рисков.

---

### 3.5. Dividend business

#### Контекст

- бизнес зрелый и прибыльный;
- фокус — стабильная прибыль, эффективность и cash flow;
- рост может быть не главным приоритетом;
- системы часто имеют длинную историю и накопленный technical debt;
- стоимость миграции и замены может быть выше потенциальной выгоды.

#### Главный инженерный фокус

- надёжность;
- стоимость владения;
- снижение операционных расходов;
- поддерживаемость;
- управляемое обновление без потери стабильности.

#### Что особенно важно

- TCO и ROI в RFC;
- стоимость простоя;
- capacity и resource efficiency;
- простые и поддерживаемые решения;
- план постепенной модернизации;
- честное описание того, что менять не нужно.

#### Риск неправильной зрелости

Команда применяет startup-подход к критичной зрелой системе или бесконечно модернизирует то, что уже достаточно хорошо работает.

---

## 4. Матрица глубины инженерных практик

| Практика | Pre-seed / Seed | Series A | Scale-up | Corporation | Dividend business |
|---|---|---|---|---|---|
| PRD | Problem, scope, user path | Metrics, NFR, risks | Dependencies, phased delivery | Stakeholders, compliance, approvals | ROI, cost of change |
| RFC | Short proposal | Alternatives and risks | Cross-team impact, migration | Governance, security, formal review | TCO, ROI, operational risk |
| ADR | Decisions that are hard to reverse | Key technology choices | Boundary and ownership decisions | Versioned decision history | Cost, supportability, longevity |
| C4 | Context + main containers | Containers + key interactions | Ownership and dependencies | Landscape, integration, governance | Legacy boundaries and change cost |
| DDD | Only useful business boundaries | Explicit ownership | Context Map and contracts | Org/domain alignment | Pragmatic boundaries around legacy |
| Data Flow | Critical path only | Data lifecycle and failure cases | Ownership, retention, consistency | Privacy, compliance, audit | Cost, retention, migration |
| Control Flow | Happy path + major failure | Retries, timeouts, idempotency | Backpressure, ordering, recovery | Formal failure and escalation model | Operational simplicity |
| Plan | Thin vertical slice; register only irreversible/high-risk debt | Milestones, risks and explicit debt capacity | Dependencies, critical path and shared debt register | Program governance, debt portfolio and approvals | ROI, TCO, debt repayment economics and capacity efficiency |
| Testing | Risk-based minimum | Test levels and regression | Contracts, performance, resilience | Evidence and release gates | Regression and stability economics |
| Release | Manual or simple rollback | Checklist and feature flag | Progressive rollout | Formal go/no-go and audit | Low-risk, reversible change |
| Incident | Roles and communication basics | Runbook for critical paths | Incident command and SLOs | Escalation and stakeholder protocol | MTTR, cost, recurrence prevention |
| Postmortem | Short learning note | Actions with owners | Systemic causes and trends | Audit and cross-org learning | Prevention ROI and operational improvement |

---

## 5. Stage-specific guidance по восьми теоретическим блокам

### Управление техническим долгом как сквозная практика

Технический долг следует рассматривать как управляемое обязательство, а не как бесконечный список «плохих задач».

Каждый debt item должен описывать:

- почему возник компромисс;
- какую стоимость он создаёт сейчас;
- какой риск создаёт в будущем;
- кто владелец;
- что станет trigger для repayment;
- сколько стоит исправление;
- какой Definition of Done закроет долг.

Классификация долга в курсе:

- architectural debt;
- code debt;
- test debt;
- infrastructure debt;
- documentation debt;
- operational debt;
- security/compliance debt.

В каждой практической системе участник отвечает на вопрос:

> Какой технический компромисс мы принимаем сейчас и как не дать ему стать невидимым?

Технический долг должен появляться в RFC/ADR, учитываться в Plan и RAID Log, проверяться при Release и разбираться в Postmortem.

### Блок 1. PRD

- **Seed:** доказать, что проблема реальна, и ограничить MVP.
- **Series A:** добавить measurable outcomes, NFR и риски.
- **Scale-up:** явно описать dependencies, ownership и rollout по командам.
- **Corporation:** учесть compliance, security, legal и согласования.
- **Dividend business:** добавить ROI, TCO, стоимость простоя и стоимость изменения.

### Блок 2. RFC / ADR

- **Seed:** коротко зафиксировать необратимые решения.
- **Series A:** сравнивать альтернативы перед началом реализации.
- **Scale-up:** показывать влияние на соседние команды и контракты.
- **Corporation:** вести formal review и decision history.
- **Dividend business:** объяснять решение через risk reduction и economics.

### Блок 3. C4

- **Seed:** не создавать архитектурную бюрократию; достаточно context/container view.
- **Series A:** сделать границы сервисов понятными команде.
- **Scale-up:** показать ownership, dependencies и integration points.
- **Corporation:** отобразить enterprise landscape и критичные внешние системы.
- **Dividend business:** визуализировать legacy, hotspots и безопасные зоны изменений.

### Блок 4. DDD

- **Seed:** выделять только границы, которые защищают бизнес-инварианты.
- **Series A:** начинать явно распределять ownership.
- **Scale-up:** строить Context Map и стабильные integration contracts.
- **Corporation:** учитывать организационные и регуляторные границы.
- **Dividend business:** не дробить систему без доказанной экономической пользы.

### Блок 5. Data Flow / Control Flow

- **Seed:** защищать критический пользовательский путь.
- **Series A:** добавить retries, timeout, idempotency и базовую наблюдаемость.
- **Scale-up:** проектировать backpressure, ordering, replay и isolation.
- **Corporation:** учитывать auditability, privacy и формальные failure procedures.
- **Dividend business:** предпочитать надёжность и простоту сложной оптимизации.

### Блок 6. Delivery planning и технический долг

- **Seed:** вертикальный slice и быстрое получение feedback; фиксировать только дорогой или необратимый долг.
- **Series A:** milestones, ownership, риски, dependencies и регулярный debt repayment capacity.
- **Scale-up:** cross-team program, critical path, общий debt register и escalation.
- **Corporation:** governance, budget, approvals, auditability и портфель технического долга.
- **Dividend business:** capacity, TCO, ROI, стоимость поддержки и минимизация disruption.

### Блок 7. Testing / Release

- **Seed:** проверять самые дорогие ошибки.
- **Series A:** строить repeatable quality process.
- **Scale-up:** contract, performance, resilience и progressive delivery.
- **Corporation:** evidence, compliance gates и formal sign-off.
- **Dividend business:** regression safety, stability и стоимость поддержки.

### Блок 8. Operations / Incidents

- **Seed:** понятно, кто реагирует и как сообщить о проблеме.
- **Series A:** runbooks для критичных операций и action-oriented postmortems.
- **Scale-up:** incident command, SLO, escalation и системные corrective actions.
- **Corporation:** formal severity model, stakeholder communication и audit trail.
- **Dividend business:** снижение MTTR, повторяемости и стоимости инцидентов.

---

## 6. Как обсуждать «слишком мало» и «слишком много»

На каждом практическом занятии ведущий должен показать два ошибочных края.

### Слишком мало

- решение держится только в голове автора;
- нет владельца;
- неизвестно, что именно проверять;
- failure scenarios не рассмотрены;
- команда узнаёт о зависимости после начала реализации.

### Слишком много

- документ создаётся без реального decision point;
- процесс не снижает конкретный риск;
- approval нужен только «потому что так принято»;
- архитектура сложнее бизнес-проблемы;
- стоимость согласования выше стоимости ошибки.

### Правильный вопрос

Не «какой документ считается правильным?», а:

> Какой минимальный артефакт позволит нужным людям принять качественное решение и потом понять, почему оно было принято?

---

## 7. Тест на достаточность документа

Документ считается достаточно глубоким, если читатель может ответить:

1. Какую проблему решаем?
2. Что именно предлагаем сделать?
3. Что сознательно не делаем?
4. Какие есть ограничения и риски?
5. Кто принимает решение?
6. Кто отвечает за реализацию и эксплуатацию?
7. Как проверим результат?
8. Что будем делать при неудаче?

Если ответить нельзя — документ недостаточно глубокий для текущего контекста.

Если ответ есть, но для его получения нужно читать 40 страниц и проводить пять встреч, документ, вероятно, избыточен.

---

## 8. Использование в курсе

В каждом теоретическом видео появляется отдельный блок:

```text
Один навык → пять стадий компании → два ошибочных края → практическая задача
```

В каждом live-семинаре ведущий может менять контекст одной и той же системы:

- «Эту же задачу проектируем для Seed»;
- «Теперь это Series A с тремя командами»;
- «Теперь это critical enterprise workflow»;
- «Теперь это зрелый dividend business с дорогой legacy-системой».

Участник учится не запоминать архитектуру, а адаптировать глубину решения к контексту.

---

## 9. Ограничения framework

- Стадия компании не определяет архитектуру автоматически.
- Размер компании не всегда равен сложности системы.
- Маленькая компания может иметь enterprise-grade compliance requirements.
- Зрелая компания может использовать lightweight process для локальной задачи.
- Dividend business не означает отсутствие инноваций.
- Любая классификация — гипотеза, которую нужно проверять по реальным ограничениям.
