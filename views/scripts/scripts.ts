type CPUPoint = {
    observedAt: string
    usedPercent: number
    coresLogical: number
    perCorePercent: number[]
    load1M: number
}

type MemoryPoint = {
    observedAt: string
    virtualUsedPercent: number
    virtualUsedBytes: number
    virtualAvailableBytes: number
    virtualTotalBytes: number
    swapUsedPercent: number
    swapUsedBytes: number
    swapAvailableBytes: number
    swapTotalBytes: number
}

type DiskPoint = {
    observedAt: string
    isTotal: boolean
    mount: string
    device: string
    filesystem: string
    usedPercent: number
    usedBytes: number
    freeBytes: number
    totalBytes: number
}

const percent = (value: number) => `${Math.round(value)}%`

function formatBytes(bytes: number): string {
    const unit = 1024
    if (bytes < unit) return `${bytes} B`

    let value = bytes
    for (const suffix of ['KB', 'MB', 'GB', 'TB', 'PB']) {
        value = value / unit
        if (value < unit) return `${value.toFixed(1)} ${suffix}`
    }

    return `${(value / unit).toFixed(1)} EB`
}

function parsePoints<T>(value: string | null): T[] {
    if (!value) return []
    return JSON.parse(value) as T[]
}

function latest<T>(points: T[]): T | undefined {
    return points.at(-1)
}

function drawHistory(canvas: HTMLCanvasElement, values: number[], stroke: string, fill: string) {
    const rect = canvas.getBoundingClientRect()
    const ratio = window.devicePixelRatio || 1
    canvas.width = rect.width * ratio
    canvas.height = rect.height * ratio

    const ctx = canvas.getContext('2d')
    if (!ctx) return

    ctx.scale(ratio, ratio)
    ctx.clearRect(0, 0, rect.width, rect.height)

    for (const mark of [25, 50, 75]) {
        const y = rect.height - (mark / 100) * rect.height
        ctx.fillStyle = 'rgba(212, 212, 216, 0.58)'
        ctx.font = '11px ui-monospace, SFMono-Regular, Menlo, monospace'
        ctx.fillText(`${mark}%`, 8, y - 4)
    }

    if (values.length === 0) return

    const xFor = (index: number) => (index / Math.max(1, values.length - 1)) * rect.width
    const yFor = (value: number) =>
        rect.height - (Math.max(0, Math.min(100, value)) / 100) * rect.height

    ctx.beginPath()
    values.forEach((value, index) => {
        const x = xFor(index)
        const y = yFor(value)
        if (index === 0) ctx.moveTo(x, y)
        else ctx.lineTo(x, y)
    })
    ctx.lineWidth = 2
    ctx.strokeStyle = stroke
    ctx.stroke()

    ctx.lineTo(rect.width, rect.height)
    ctx.lineTo(0, rect.height)
    ctx.closePath()

    const gradient = ctx.createLinearGradient(0, 0, 0, rect.height)
    gradient.addColorStop(0, fill)
    gradient.addColorStop(1, 'rgba(9, 9, 11, 0.05)')
    ctx.fillStyle = gradient
    ctx.fill()
}

function setMeterWidth(root: ParentNode, selector: string, value: number) {
    const meter = root.querySelector<HTMLElement>(selector)
    if (meter) meter.style.width = `${Math.max(0, Math.min(100, value))}%`
}

class OpenLogsCPUHistory extends HTMLElement {
    static observedAttributes = ['data-points']
    private resize = () => this.draw()

    connectedCallback() {
        window.addEventListener('resize', this.resize)
        this.render()
    }

    disconnectedCallback() {
        window.removeEventListener('resize', this.resize)
    }

    attributeChangedCallback() {
        this.render()
    }

    private points() {
        return parsePoints<CPUPoint>(this.getAttribute('data-points'))
    }

    private render() {
        const points = this.points()
        const current = latest(points)
        const load = current
            ? `load ${current.load1M.toFixed(2)} . ${current.coresLogical} logical cores`
            : 'Waiting for CPU history'

        this.innerHTML = `
            <div class="flex flex-col gap-1 sm:flex-row sm:items-start sm:justify-between">
                <div>
                    <h2 class="text-lg font-semibold">CPU history</h2>
                    <p class="text-sm text-zinc-500">${load}</p>
                </div>
                <div class="flex items-center gap-2 text-xs text-zinc-500">
                    <span class="inline-flex items-center gap-1"><span class="h-2 w-2 bg-emerald-400"></span>low</span>
                    <span class="inline-flex items-center gap-1"><span class="h-2 w-2 bg-amber-400"></span>warm</span>
                    <span class="inline-flex items-center gap-1"><span class="h-2 w-2 bg-red-400"></span>hot</span>
                </div>
            </div>
            <div class="graph-grid mt-4 h-64 border border-zinc-700 bg-zinc-950">
                <canvas class="history-canvas" aria-label="CPU utilization history"></canvas>
            </div>
            <div class="core-rows mt-4 space-y-2"></div>
        `

        this.renderCoreRows(points)
        requestAnimationFrame(() => this.draw())
    }

    private renderCoreRows(points: CPUPoint[]) {
        const root = this.querySelector<HTMLElement>('.core-rows')
        if (!root) return

        const current = latest(points)
        const coreCount = current?.perCorePercent.length || current?.coresLogical || 0
        if (coreCount === 0) {
            root.innerHTML = `<div class="text-sm text-zinc-500">Waiting for per-core samples</div>`
            return
        }

        root.innerHTML = ''
        for (let core = 0; core < coreCount; core++) {
            const values = points
                .map((point) => point.perCorePercent[core] ?? point.usedPercent)
                .slice(-72)
            const lastValue = values.at(-1) ?? 0
            const row = document.createElement('div')
            row.className = 'grid grid-cols-[3.5rem_minmax(0,1fr)_3rem] items-center gap-2 text-xs'
            row.innerHTML = `
                <span class="text-zinc-500">cpu${core}</span>
                <div class="flex h-4 gap-px overflow-hidden"></div>
                <span class="text-right text-zinc-300">${percent(lastValue)}</span>
            `

            const meter = row.children[1] as HTMLElement
            for (const value of values) {
                const cell = document.createElement('span')
                cell.className = 'meter-cell'
                if (value > 80) cell.classList.add('on-high')
                else if (value > 60) cell.classList.add('on-mid')
                else if (value > 5) cell.classList.add('on-low')
                meter.appendChild(cell)
            }

            root.appendChild(row)
        }
    }

    private draw() {
        const canvas = this.querySelector<HTMLCanvasElement>('canvas')
        if (!canvas) return

        drawHistory(
            canvas,
            this.points().map((point) => point.usedPercent),
            'rgb(52, 211, 153)',
            'rgba(52, 211, 153, 0.42)',
        )
    }
}

class OpenLogsMemoryPressure extends HTMLElement {
    static observedAttributes = ['data-points']
    private resize = () => this.draw()

    connectedCallback() {
        window.addEventListener('resize', this.resize)
        this.render()
    }

    disconnectedCallback() {
        window.removeEventListener('resize', this.resize)
    }

    attributeChangedCallback() {
        this.render()
    }

    private points() {
        return parsePoints<MemoryPoint>(this.getAttribute('data-points'))
    }

    private render() {
        const points = this.points()
        const current = latest(points)
        const badge = current && current.virtualUsedPercent >= 85 ? 'hot' : 'steady'
        const ramText = current
            ? `${formatBytes(current.virtualUsedBytes)} used . ${formatBytes(current.virtualAvailableBytes)} free`
            : 'Waiting for RAM samples'
        const swapText = current
            ? `${formatBytes(current.swapUsedBytes)} used . ${formatBytes(current.swapAvailableBytes)} free`
            : 'Waiting for swap samples'

        this.innerHTML = `
            <div class="flex items-start justify-between gap-4">
                <div>
                    <h2 class="text-lg font-semibold">Memory pressure</h2>
                    <p class="text-sm text-zinc-500">Used memory and swap pressure.</p>
                </div>
                <span class="rounded bg-amber-500 px-2 py-1 text-xs font-semibold text-zinc-950">${badge}</span>
            </div>
            <div class="graph-grid mt-4 h-64 border border-zinc-700 bg-zinc-950">
                <canvas class="history-canvas" aria-label="Memory usage history"></canvas>
            </div>
            <div class="mt-4 space-y-3 text-sm">
                <div>
                    <div class="mb-1 flex justify-between gap-4 text-zinc-400">
                        <span>RAM</span>
                        <span>${ramText}</span>
                    </div>
                    <div class="h-3 overflow-hidden bg-zinc-800"><div class="ram-meter h-full bg-amber-400"></div></div>
                </div>
                <div>
                    <div class="mb-1 flex justify-between gap-4 text-zinc-400">
                        <span>Swap</span>
                        <span>${swapText}</span>
                    </div>
                    <div class="h-3 overflow-hidden bg-zinc-800"><div class="swap-meter h-full bg-cyan-400"></div></div>
                </div>
            </div>
        `

        setMeterWidth(this, '.ram-meter', current?.virtualUsedPercent ?? 0)
        setMeterWidth(this, '.swap-meter', current?.swapUsedPercent ?? 0)
        requestAnimationFrame(() => this.draw())
    }

    private draw() {
        const canvas = this.querySelector<HTMLCanvasElement>('canvas')
        if (!canvas) return

        drawHistory(
            canvas,
            this.points().map((point) => point.virtualUsedPercent),
            'rgb(251, 191, 36)',
            'rgba(251, 191, 36, 0.48)',
        )
    }
}

class OpenLogsDiskPressure extends HTMLElement {
    static observedAttributes = ['data-points']

    connectedCallback() {
        this.render()
    }

    attributeChangedCallback() {
        this.render()
    }

    private points() {
        return parsePoints<DiskPoint>(this.getAttribute('data-points'))
    }

    private render() {
        const disks = this.latestDisks()
        const worst = disks[0]
        const badge = worst
            ? `${worst.mount || 'disk'} ${worst.usedPercent >= 85 ? 'needs attention' : 'healthy'}`
            : 'waiting'

        this.innerHTML = `
            <div class="flex flex-col gap-1 sm:flex-row sm:items-start sm:justify-between">
                <div>
                    <h2 class="text-lg font-semibold">Disk pressure</h2>
                    <p class="text-sm text-zinc-500">Sorted by least free space.</p>
                </div>
                <span class="rounded bg-red-500 px-2 py-1 text-xs font-semibold text-white">${badge}</span>
            </div>
            <div class="mt-4 overflow-x-auto">
                <table class="w-full text-left text-sm">
                    <thead>
                        <tr class="text-zinc-500">
                            <th class="border-b border-zinc-800 px-4 py-2">mount</th>
                            <th class="border-b border-zinc-800 px-4 py-2">fs</th>
                            <th class="border-b border-zinc-800 px-4 py-2">used</th>
                            <th class="border-b border-zinc-800 px-4 py-2">free</th>
                            <th class="border-b border-zinc-800 px-4 py-2">history</th>
                        </tr>
                    </thead>
                    <tbody>
                        ${disks.length === 0 ? this.emptyRow() : disks.map((disk) => this.diskRow(disk)).join('')}
                    </tbody>
                </table>
            </div>
        `

        for (const bar of this.querySelectorAll<HTMLElement>('[data-disk-width]')) {
            bar.style.width = `${bar.dataset.diskWidth}%`
        }
    }

    private latestDisks() {
        const byMount = new Map<string, DiskPoint>()
        for (const point of this.points()) {
            if (point.isTotal) continue
            byMount.set(point.mount || point.device || 'disk', point)
        }

        return [...byMount.values()].sort((a, b) => {
            const aFree = 100 - a.usedPercent
            const bFree = 100 - b.usedPercent
            return aFree - bFree
        })
    }

    private emptyRow() {
        return `<tr><td class="px-4 py-6 text-zinc-500" colspan="5">Waiting for disk samples</td></tr>`
    }

    private diskRow(disk: DiskPoint) {
        const tone =
            disk.usedPercent >= 85
                ? 'text-red-400'
                : disk.usedPercent >= 65
                  ? 'text-amber-400'
                  : 'text-emerald-400'
        const fill =
            disk.usedPercent >= 85
                ? 'bg-red-400'
                : disk.usedPercent >= 65
                  ? 'bg-amber-400'
                  : 'bg-emerald-400'
        const width = Math.max(0, Math.min(100, disk.usedPercent))

        return `
            <tr>
                <th class="border-b border-zinc-900 px-4 py-3 font-semibold">${disk.mount || disk.device || '-'}</th>
                <td class="border-b border-zinc-900 px-4 py-3 text-zinc-300">${disk.filesystem || '-'}</td>
                <td class="border-b border-zinc-900 px-4 py-3 ${tone}">${percent(disk.usedPercent)}</td>
                <td class="border-b border-zinc-900 px-4 py-3 text-zinc-300">${formatBytes(disk.freeBytes)}</td>
                <td class="border-b border-zinc-900 px-4 py-3">
                    <div class="h-2 w-40 bg-zinc-800">
                        <div class="h-full ${fill}" data-disk-width="${width}"></div>
                    </div>
                </td>
            </tr>
        `
    }
}

if (!customElements.get('openlogs-cpu-history')) {
    customElements.define('openlogs-cpu-history', OpenLogsCPUHistory)
}

if (!customElements.get('openlogs-memory-pressure')) {
    customElements.define('openlogs-memory-pressure', OpenLogsMemoryPressure)
}

if (!customElements.get('openlogs-disk-pressure')) {
    customElements.define('openlogs-disk-pressure', OpenLogsDiskPressure)
}
