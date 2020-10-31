// Copyright 2020 The casbin Authors. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

import { message } from 'antd';
import { isMobile as isMobileDevice } from 'react-device-detect';

export function parseJson(s: string) {
  if (s === '') {
    return null;
  } else {
    return JSON.parse(s);
  }
}

export function myParseInt(i: string) {
  const res = parseInt(i);
  return isNaN(res) ? 0 : res;
}

export function openLink(link: string) {
  const w = window.open('about:blank');
  if (w) {
    w.location.href = link;
  }
}

export function goToLink(link: string) {
  window.location.href = link;
}

export function showMessage(type: '' | 'success' | 'error', text: string) {
  if (type === '') {
    return;
  } else if (type === 'success') {
    message.success(text);
  } else if (type === 'error') {
    message.error(text);
  }
}

export function deepCopy(obj: object) {
  return Object.assign({}, obj);
}

export function addRow(array: Array<any>, row: any) {
  return [...array, row];
}

export function prependRow(array: Array<any>, row: any) {
  return [row, ...array];
}

export function deleteRow(array: Array<any>, i: number) {
  // return array = array.slice(0, i).concat(array.slice(i + 1));
  return [...array.slice(0, i), ...array.slice(i + 1)];
}

export function swapRow(array: Array<any>, i: number, j: number) {
  return [...array.slice(0, i), array[j], ...array.slice(i + 1, j), array[i], ...array.slice(j + 1)];
}

export function isMobile() {
  return isMobileDevice;
}

export function getFormattedDate(date: string | undefined) {
  if (date === undefined) {
    return null;
  }

  date = date.replace('T', ' ');
  date = date.replace('+08:00', ' ');
  return date;
}

export function getFormattedDateShort(date: string) {
  return date.slice(0, 10);
}

export function getShortName(s: string) {
  return s.split('/').slice(-1)[0];
}

function getRandomInt(s: string) {
  let hash = 0;
  if (s.length !== 0) {
    for (let i = 0; i < s.length; i++) {
      let char = s.charCodeAt(i);
      hash = (hash << 5) - hash + char;
      hash = hash & hash;
    }
  }

  return hash;
}

export function getAvatarColor(s: string) {
  const colorList = ['#f56a00', '#7265e6', '#ffbf00', '#00a2ae'];
  let random = getRandomInt(s);
  if (random < 0) {
    random = -random;
  }
  return colorList[random % 4];
}
