import styled, { keyframes } from 'styled-components/macro';

const glow = keyframes`
from, to {
  color: black;
}

50% {
  color: rgba(225,0,0,.8);
}
`;

export const Num = styled.span`
  font-variant-numeric: tabular-nums;
  &.glow {
    animation: ${glow} 500ms;
  }
`;

export const Main = styled.section`
  display        : flex;
  flex-direction : column;
  justify-content: center;
  align-items    : center;

  p {
    margin-top   : 30px;
    overflow     : hidden;
    font-size    : 30px;
    font-weight  : 700;
    text-overflow: ellipsis;
    white-space  : nowrap;
  }
`;

export const Logo = styled.img`
  margin-top: 100px;
  width     : 100px;
`;
