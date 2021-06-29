import React, { useState, useEffect } from 'react';
import RealTimeQuery, { createWebSocketTransport, rxjsOperators } from 'real-time-query';
import cx from 'classnames';
import { Main, Logo, Num } from './ui';
import base64 from 'react-native-base64';

const WEB_SOCKET_URL = process.env.REACT_APP_WEB_SOCKET_URL === undefined ? 'http://localhost:8000' : process.env.REACT_APP_WEB_SOCKET_URL
const WEB_SOCKET_PATH = process.env.REACT_APP_WEB_SOCKET_PATH === undefined ? '/socket.io' : process.env.REACT_APP_WEB_SOCKET_PATH

const OPYTIONS = process.env.NODE_ENV === 'production'
    ? { apiUrl: WEB_SOCKET_URL, path: WEB_SOCKET_PATH }
    : { apiUrl: WEB_SOCKET_URL, path: WEB_SOCKET_PATH };

export default function App() {
  const [result, setResult] = useState(null);

  useEffect(() => {
    const realTimeQuery = new RealTimeQuery({
      transport: createWebSocketTransport(OPYTIONS)
    });

    const { pairwise, timestamp } = rxjsOperators;

    realTimeQuery.subscribe(
        {
          eventName: 'receive_sink',
          rxjsOperators: [
            pairwise(),
            timestamp(),
          ]
        },
        result => {
          setResult(result);
        }
    );

    return () => {
      realTimeQuery.close();
    }
  }, []);

  if (!result) {
    return null;
  }

  return (
      <Main>
        <Logo className='logo' src='logo.png' alt='YoMo' />
        <p>
          Real-time shake level:
          <Num className={cx({ glow: result.value[0].payload !== result.value[1].payload })}>
              {base64.decode(result.value[1].payload)}
          </Num>
        </p>
        <span>Delay: <Num>{result.timestamp - result.value[1].time}ms</Num></span>
      </Main>
  )
};