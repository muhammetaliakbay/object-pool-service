import WebSocket = require("ws")
import { promisify } from "util"
import { Observable, Subject } from "rxjs"

export namespace ObjectPool {
    export interface Mark {
        mark: number,
        queue: {
            [group: string]: string[],
        },
    }

    export type Result = {
        release?: string[],
    } | undefined

    export type Loader = (args: {size: number}) => Mark | Promise<Mark>
    export type Processor = (args: {objects: string[]}) => Result | Promise<Result>

    export type Disjoin = () => void

    class SessionImpl implements Session {
        constructor(
            private socket: WebSocket,
            private loader: Loader,
            private processor: Processor,
        ) {
        }
        private send = promisify(this.socket.send.bind(this.socket))
        private disjoinSubject = new Subject<never>()
        readonly onDisjoin = this.disjoinSubject.asObservable()

        disjoin() {
            this.disjoinSubject.complete()
        }

        listen() {
            const onMessage = (data: WebSocket.RawData) => {
                let json: any
                try {
                    json = JSON.parse(data.toString("utf-8"))
                } catch (err) {
                    this.disjoin()
                    return
                }
                if(typeof json !== 'object') {
                    this.disjoin()
                    return
                }
                const type = json.type
                if (typeof type !== 'string') {
                    this.disjoin()
                    return
                }
                switch(type) {
                    case 'claim':
                        const objects = json.objects
                        if (!Array.isArray(objects) || objects.some(object => typeof object !== 'string')) {
                            this.disjoin()
                            return
                        }
                        this.onClaim(objects as string[])
                        break
                    case 'load':
                        const size = json.size;
                        if (typeof size !== 'number' || !(size >= 0) || Math.floor(size) !== size) {
                            this.disjoin()
                            return
                        }
                        this.onLoad(size)
                        break
                }
            }
            this.socket.on('message', onMessage)

            const onClose = () => {
                this.disjoinSubject.complete()
            }
            this.socket.once('close', onClose)

            this.disjoinSubject.subscribe({
                complete: () => {
                    this.socket.off('message', onMessage)
                    this.socket.off('close', onClose)
                }
            })
        }

        private async onClaim(objects: string[]) {
            try {
                const { release = objects } = await this.processor({
                    objects,
                }) ?? {}
                if (release.length > 0) {
                    await this.send(JSON.stringify({
                        type: 'release',
                        objects: release,
                    }))
                }
                const requeue = objects.filter(
                    object => release.includes(object),
                )
                await this.send(JSON.stringify({
                    type: 'requeue',
                    objects: requeue,
                }))
            } catch (err) {
                this.disjoin()
            }
        }

        private async onLoad(size: number) {
            try {
                const result = await this.loader({
                    size,
                })
                for (const [group, objects] of Object.entries(result.queue)) {
                    await this.send(JSON.stringify({
                        type: 'queue',
                        group,
                        objects,
                    }))
                }
                await this.send(JSON.stringify({
                    type: 'mark',
                    size: result.mark,
                }))
            } catch (err) {
                this.disjoin()
            }
        }
    }

    export interface Session {
        readonly onDisjoin: Observable<never>
        disjoin(): void
    }

    export async function join({
        url: baseURL,
        pool,
        limit,
        loader,
        processor,
    }: {
        url: string | URL,
        pool: string,
        limit: number,
        loader: Loader,
        processor: Processor,
    }): Promise<Session> {
        const url = new URL(baseURL)
        url.searchParams.append("pool", pool)
        url.searchParams.append("limit", limit.toString())
        const socket = new WebSocket(url)
        await new Promise(
            (resolve, reject) => {
                const cleanup = () => {
                    socket.off('error', onError)
                    socket.off('open', onOpen)
                }
                const onError = (error: Error) => {
                    reject(error)
                    cleanup()
                }
                const onOpen = () => {
                    resolve(null)
                    cleanup()
                }
                socket.once('error', onError)
                socket.once('open', onOpen)
            }
        )
        
        const session = new SessionImpl(socket, loader, processor)
        session.listen()

        return session
    }
}

export default ObjectPool;
