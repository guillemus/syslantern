type HistoryPoint = Record<string, number | string | boolean | number[]>

function parsePoints(value: string | null): HistoryPoint[] {
    if (!value) return []
    return JSON.parse(value) as HistoryPoint[]
}

function drawHistory(
    canvas: HTMLCanvasElement,
    width: number,
    height: number,
    values: number[],
    stroke: string,
    fill: string,
) {
    const ratio = window.devicePixelRatio || 1
    canvas.width = width * ratio
    canvas.height = height * ratio

    const ctx = canvas.getContext('2d')
    if (!ctx) return

    ctx.scale(ratio, ratio)
    ctx.clearRect(0, 0, width, height)

    for (const mark of [25, 50, 75]) {
        const y = height - (mark / 100) * height
        ctx.fillStyle = 'rgba(212, 212, 216, 0.58)'
        ctx.font = '11px ui-monospace, SFMono-Regular, Menlo, monospace'
        ctx.fillText(`${mark}%`, 8, y - 4)
    }

    if (values.length === 0) return

    const xFor = (index: number) => (index / Math.max(1, values.length - 1)) * width
    const yFor = (value: number) => height - (Math.max(0, Math.min(100, value)) / 100) * height

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

    ctx.lineTo(width, height)
    ctx.lineTo(0, height)
    ctx.closePath()

    const gradient = ctx.createLinearGradient(0, 0, 0, height)
    gradient.addColorStop(0, fill)
    gradient.addColorStop(1, 'rgba(9, 9, 11, 0.05)')
    ctx.fillStyle = gradient
    ctx.fill()
}

class SyslanternHistoryCanvas extends HTMLElement {
    static observedAttributes = ['data-points', 'data-value-key', 'data-stroke', 'data-fill']
    private readonly root: ShadowRoot
    private readonly canvas: HTMLCanvasElement
    private readonly resizeObserver = new ResizeObserver(() => this.draw())

    constructor() {
        super()
        this.root = this.attachShadow({ mode: 'open' })
        this.canvas = document.createElement('canvas')
        this.canvas.style.display = 'block'
        this.canvas.style.width = '100%'
        this.canvas.style.height = '100%'
    }

    connectedCallback() {
        if (this.canvas.parentNode !== this.root) this.root.appendChild(this.canvas)
        this.resizeObserver.observe(this)
        requestAnimationFrame(() => this.draw())
    }

    disconnectedCallback() {
        this.resizeObserver.unobserve(this)
    }

    attributeChangedCallback() {
        if (this.isConnected) requestAnimationFrame(() => this.draw())
    }

    private draw() {
        const rect = this.getBoundingClientRect()
        if (rect.width === 0 || rect.height === 0) return

        const key = this.dataset.valueKey || 'value'
        const values = parsePoints(this.dataset.points || null)
            .map((point) => point[key])
            .filter((value): value is number => typeof value === 'number')

        drawHistory(
            this.canvas,
            rect.width,
            rect.height,
            values,
            this.dataset.stroke || 'rgb(52, 211, 153)',
            this.dataset.fill || 'rgba(52, 211, 153, 0.42)',
        )
    }
}

if (!customElements.get('syslantern-history-canvas')) {
    customElements.define('syslantern-history-canvas', SyslanternHistoryCanvas)
}
