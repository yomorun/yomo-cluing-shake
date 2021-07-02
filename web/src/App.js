import React, { useState, useEffect } from 'react';
import RealTimeQuery, { createWebSocketTransport, rxjsOperators } from 'real-time-query';
import cx from 'classnames';
import { Main, Logo, Num } from './ui';
//import base64 from 'react-native-base64';

const WEB_SOCKET_URL = process.env.REACT_APP_WEB_SOCKET_URL === undefined ? 'http://localhost:8000' : process.env.REACT_APP_WEB_SOCKET_URL
const WEB_SOCKET_PATH = process.env.REACT_APP_WEB_SOCKET_PATH === undefined ? '/socket.io' : process.env.REACT_APP_WEB_SOCKET_PATH

const OPYTIONS = process.env.NODE_ENV === 'production'
    ? { apiUrl: WEB_SOCKET_URL, path: WEB_SOCKET_PATH }
    : { apiUrl: WEB_SOCKET_URL, path: WEB_SOCKET_PATH };

console.debug(process.env)
console.debug(OPYTIONS)

export default function App() {
  const [resultS07, setResultS07] = useState(null);
  const [resultS05, setResultS05] = useState(null);

    useEffect(() => {
        const realTimeQuery = new RealTimeQuery({
            transport: createWebSocketTransport(OPYTIONS)
        });

        const { pairwise, timestamp } = rxjsOperators;

        realTimeQuery.subscribe(
            {
                eventName: 'receive_sink_s07',
                rxjsOperators: [
                    pairwise(),
                    timestamp(),
                ]
            },
            result => {
                setResultS07(result);
            }
        );

        realTimeQuery.subscribe(
            {
                eventName: 'receive_sink_s05',
                rxjsOperators: [
                    pairwise(),
                    timestamp(),
                ]
            },
            result => {
                setResultS05(result);
            }
        );

        return () => {
            realTimeQuery.close();
        }
    }, []);

  return (
      <Main>
        <Logo className='logo' src='logo.png' alt='YoMo' />
          {
              resultS07 && (
                  <p>
                      TOPIC: {resultS07.value[0].topic}
                      <li>
                          temperature:
                          <Num className={cx({ glow: resultS07.value[0].temperature === resultS07.value[1].temperature })}>
                              {resultS07.value[1].temperature}
                          </Num>
                      </li>
                      <li>
                          vertical:
                          <Num className={cx({ glow: resultS07.value[0].vertical === resultS07.value[1].vertical })}>
                              {resultS07.value[1].vertical}
                          </Num>
                      </li>
                      <li>
                          transverse:
                          <Num className={cx({ glow: resultS07.value[0].transverse === resultS07.value[1].transverse })}>
                              {resultS07.value[1].transverse}
                          </Num>
                      </li>
                      <br/>
                      <span>Delay: <Num>{resultS07.timestamp - resultS07.value[1].time}ms</Num></span>
                  </p>
              )
          }

          {
              resultS05 && (
                  <p>
                      TOPIC: {resultS05.value[0].topic}
                      <li>
                          Key:
                          <Num className={cx({ glow: resultS05.value[0].key === resultS05.value[1].key })}>
                              {resultS05.value[1].key}
                          </Num>
                      </li>
                      <br/>
                      <span>Delay: <Num>{resultS05.timestamp - resultS05.value[1].time}ms</Num></span>
                  </p>
              )
          }

      </Main>
  )
};