import { TypedUseSelectorHook, useDispatch, useSelector } from 'react-redux';
import type { RootState, AppDispatch } from './configureStore';

/**
 * 使用型別安全的 hooks 而非原生的 useDispatch 和 useSelector
 * 這樣可以獲得更好的 TypeScript 支援和自動補全
 */
export const useAppDispatch = () => useDispatch<AppDispatch>();
export const useAppSelector: TypedUseSelectorHook<RootState> = useSelector;