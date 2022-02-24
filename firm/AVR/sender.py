#!/usr/bin/python3

# apt install python3-aiohttp python3-serial-asyncio
import asyncio,serial_asyncio,aiohttp
import hashlib,logging,time

shareKey = '_SHARE_KEY_'
target = 'http://example.com/do'
place = 'tokyo'
ttyDev = '/dev/ttyAMA0'

ambientWriteKey = 'HOGEHOGE'
ambientChannelId = 114514

class ACFreqWatcher(asyncio.Protocol):
    def __init__ (self):
        # If you initializes at out of __init__ , You will receive an exception 'attached to a different loop'
        self.__queue = asyncio.Queue()

    def connection_made(self, transport):
        self.transport = transport
        logging.info('port opened :%s', transport)

    __data = b''

    def data_received(self, data):
        self.__data += data
        i = self.__data.find(b'\r\n')
        if i < 0:
            return
        chunk = self.__data[:i]
        self.__data = self.__data[i+2:]
        self.__freq_received(chunk.decode('us-ascii'))

    def __freq_received(self,data):
        freqs = data.split(" ")
        logging.debug("data received :%s",freqs)

        self.__queue.put_nowait({
                'place':place,
                'freq': float(freqs[1]) / 10000.
            })

    def get_queue(self):
        return self.__queue

class ACFreqSender:
    __metric_timeout = aiohttp.ClientTimeout(total=0.65)
    __ambient_timeout = aiohttp.ClientTimeout(total=10)
    __last_ambient_sent = 0

    async def __send_metric_async(self,session:aiohttp.ClientSession,place:str,freq:float) -> bool:
        body = place + str(freq) + shareKey
        sign = hashlib.md5(body.encode('us-ascii')).hexdigest()

        # args must be int or str
        args = {'place':place,'freq':str(freq),'sign':sign}
        
        try:
            async with session.get(target,params=args,timeout=self.__metric_timeout) as res:
                res.raise_for_status()
                logging.debug("sent data to %s :%s",target,args)
                return True
        except Exception as e:
                logging.warning("HTTP Error:%s",repr(e))
                return False

    async def __send_ambient_async(self,session:aiohttp.ClientSession,freq:float) -> bool:
        now = int(time.time())
        if now <= self.__last_ambient_sent + 30:
            return False
        self.__last_ambient_sent = now

        obj = {
                    'writeKey': ambientWriteKey,
                    'data':[
                        {'d1': freq,}
                    ],
                }
        try:
            async with session.post('http://ambidata.io/api/v2/channels/{}/dataarray'.format(ambientChannelId),
                json=obj,timeout=self.__ambient_timeout) as res:
                res.raise_for_status()
                logging.debug("sent ambient to %s :%s",target,obj)
                return True
        except Exception as e:
                logging.warning("HTTP(Ambient) Error:%s",repr(e))
                return False
    

    async def send_metric_from_queue_async(self,queue:asyncio.Queue) -> None:
        async with aiohttp.ClientSession() as session:
            while True:
                it = await queue.get()
                # fire and forget
                asyncio.create_task(self.__send_metric_async(session,it['place'],it['freq']))
                asyncio.create_task(self.__send_ambient_async(session,it['freq']))
                queue.task_done()

async def main():
    uartask = serial_asyncio.create_serial_connection(asyncio.get_event_loop(), ACFreqWatcher, ttyDev, baudrate=38400)
    _ , watcher = await uartask
    sendertask = ACFreqSender().send_metric_from_queue_async(watcher.get_queue())

    # never return.
    await asyncio.gather(uartask,sendertask)

if __name__ == '__main__':
    logger = logging.getLogger(__name__)
    fmt = "%(asctime)s %(levelname)s :%(message)s"
    logging.basicConfig(level=logging.INFO, format=fmt)
    asyncio.run(main())
