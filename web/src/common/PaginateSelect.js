// Copyright 2026 The Casdoor Authors. All Rights Reserved.
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

import React from "react";
import {Select, Spin} from "antd";
import * as Setting from "../Setting";

const SCROLL_BOTTOM_OFFSET = 20;

const defaultOptionMapper = (item) => {
  if (item === null) {
    return null;
  }
  if (typeof item === "string") {
    return Setting.getOption(item, item);
  }
  const value = item.value ?? item.name ?? item.id ?? item.key;
  const label = item.label ?? item.displayName ?? value;
  if (value === undefined) {
    return null;
  }
  return Setting.getOption(label, value);
};

function PaginateSelect(props) {
  const {
    fetchPage,
    buildFetchArgs,
    optionMapper = defaultOptionMapper,
    pageSize = Setting.MAX_PAGE_SIZE,
    debounceMs = Setting.SEARCH_DEBOUNCE_MS,
    onError,
    onSearch: onSearchProp,
    onPopupScroll: onPopupScrollProp,
    showSearch = true,
    filterOption = false,
    notFoundContent,
    loading: selectLoading,
    dropdownMatchSelectWidth = false,
    virtual = false,
    reloadKey,
    ...restProps
  } = props;

  const [options, setOptions] = React.useState([]);
  const [hasMore, setHasMore] = React.useState(true);
  const [loading, setLoading] = React.useState(false);

  const debounceRef = React.useRef(null);
  const latestSearchRef = React.useRef("");
  const loadingRef = React.useRef(false);
  const requestIdRef = React.useRef(0);
  const pageRef = React.useRef(0);
  const fetchPageRef = React.useRef(fetchPage);
  const buildFetchArgsRef = React.useRef(buildFetchArgs);
  const optionMapperRef = React.useRef(optionMapper ?? defaultOptionMapper);

  React.useEffect(() => {
    fetchPageRef.current = fetchPage;
  }, [fetchPage]);

  React.useEffect(() => {
    buildFetchArgsRef.current = buildFetchArgs;
  }, [buildFetchArgs]);

  React.useEffect(() => {
    optionMapperRef.current = optionMapper ?? defaultOptionMapper;
  }, [optionMapper]);

  const handleError = React.useCallback((error) => {
    if (onError) {
      onError(error);
      return;
    }

    if (Setting?.showMessage) {
      Setting.showMessage("error", error?.message ?? String(error));
    }
  }, [onError]);

  const extractItems = React.useCallback((response) => {
    if (Array.isArray(response)) {
      return response;
    }
    if (Array.isArray(response?.items)) {
      return response.items;
    }
    if (Array.isArray(response?.data)) {
      return response.data;
    }
    if (Array.isArray(response?.list)) {
      return response.list;
    }
    return [];
  }, []);

  const mergeOptions = React.useCallback((prev, next, reset) => {
    if (reset) {
      return next;
    }

    const merged = [...prev];
    const indexByValue = new Map();
    merged.forEach((opt, idx) => {
      if (opt?.value !== undefined) {
        indexByValue.set(opt.value, idx);
      }
    });

    next.forEach((opt) => {
      if (!opt) {
        return;
      }
      const optionValue = opt.value;
      if (optionValue === undefined) {
        merged.push(opt);
        return;
      }
      if (indexByValue.has(optionValue)) {
        merged[indexByValue.get(optionValue)] = opt;
        return;
      }
      indexByValue.set(optionValue, merged.length);
      merged.push(opt);
    });

    return merged;
  }, []);

  const loadPage = React.useCallback(async({pageToLoad = 1, reset = false, search = latestSearchRef.current} = {}) => {
    const fetcher = fetchPageRef.current;
    if (typeof fetcher !== "function") {
      return;
    }
    if (loadingRef.current && !reset) {
      return;
    }
    if (reset) {
      loadingRef.current = false;
    }

    const currentRequestId = requestIdRef.current + 1;
    requestIdRef.current = currentRequestId;

    loadingRef.current = true;
    setLoading(true);

    const defaultArgsObject = {
      page: pageToLoad,
      pageSize,
      search,
      searchText: search,
      query: search,
    };

    try {
      const argsBuilder = buildFetchArgsRef.current;
      const builtArgs = argsBuilder ? argsBuilder({
        page: pageToLoad,
        pageSize,
        searchText: search,
      }) : defaultArgsObject;

      const payload = Array.isArray(builtArgs) ?
        await fetcher(...builtArgs) :
        await fetcher(builtArgs ?? defaultArgsObject);

      if (currentRequestId !== requestIdRef.current) {
        return;
      }

      if (payload?.status && payload.status !== "ok") {
        handleError(payload?.msg ?? payload?.error ?? "Request failed");
        setHasMore(false);
        return;
      }

      const items = extractItems(payload);
      const mapper = optionMapperRef.current ?? defaultOptionMapper;
      const mappedOptions = items.map(mapper).filter(Boolean);
      setOptions((prev) => mergeOptions(prev, mappedOptions, reset));
      pageRef.current = pageToLoad;

      const hasMoreFromPayload = typeof payload?.hasMore === "boolean" ? payload.hasMore : null;
      const hasMoreFromTotal = typeof payload?.total === "number" ? (pageToLoad * pageSize < payload.total) : null;
      const fallbackHasMore = mappedOptions.length === pageSize;
      setHasMore(hasMoreFromPayload ?? hasMoreFromTotal ?? fallbackHasMore);
    } catch (error) {
      if (currentRequestId === requestIdRef.current) {
        handleError(error);
      }
    } finally {
      if (currentRequestId === requestIdRef.current) {
        loadingRef.current = false;
        setLoading(false);
      }
    }
  }, [pageSize, extractItems, mergeOptions, handleError]);

  const resetAndLoad = React.useCallback((search = "") => {
    latestSearchRef.current = search;
    setOptions([]);
    setHasMore(true);
    pageRef.current = 0;
    loadPage({pageToLoad: 1, reset: true, search});
  }, [loadPage]);

  React.useEffect(() => {
    resetAndLoad("");
    return () => {
      if (debounceRef.current) {
        clearTimeout(debounceRef.current);
      }
    };
  }, [resetAndLoad, reloadKey]);

  const handleSearch = React.useCallback((value) => {
    onSearchProp?.(value);
    if (debounceRef.current) {
      clearTimeout(debounceRef.current);
    }

    const triggerSearch = () => resetAndLoad(value || "");

    if (!debounceMs) {
      triggerSearch();
      return;
    }

    debounceRef.current = setTimeout(triggerSearch, debounceMs);
  }, [debounceMs, onSearchProp, resetAndLoad]);

  const handlePopupScroll = React.useCallback((event) => {
    onPopupScrollProp?.(event);
    const target = event?.target;
    if (!target || loadingRef.current || !hasMore) {
      return;
    }

    const reachedBottom = target.scrollTop + target.offsetHeight >= target.scrollHeight - SCROLL_BOTTOM_OFFSET;
    if (reachedBottom) {
      const nextPage = pageRef.current + 1;
      loadPage({pageToLoad: nextPage});
    }
  }, [hasMore, loadPage, onPopupScrollProp]);

  const mergedLoading = selectLoading ?? loading;
  const mergedNotFound = mergedLoading ? <Spin size="small" /> : notFoundContent;

  return (
    <Select
      {...restProps}
      virtual={virtual}
      showSearch={showSearch}
      filterOption={filterOption}
      options={options}
      loading={mergedLoading}
      notFoundContent={mergedNotFound}
      onSearch={showSearch ? handleSearch : undefined}
      onPopupScroll={handlePopupScroll}
      dropdownMatchSelectWidth={dropdownMatchSelectWidth}
    />
  );
}

export default PaginateSelect;
