<template>
    <div id="yamleditordom" style="height:100%;">
        <div id="editor_textarea"></div>
        <div v-if="nofooter!==true" class="mt-16">
            <a-button type="primary" @click="submit">确定</a-button>
            <a-button class="ml-10" @click="cancel">取消</a-button>
        </div>
    </div>
</template>

<script>

import {basicSetup} from "codemirror"
// import {indentWithTab} from "@codemirror/commands"
import { EditorView } from "codemirror"
import { yaml } from "@codemirror/lang-yaml";
import { keymap } from '@codemirror/view';

export default {
    props: ['yaml','disabled','nofooter'],
    data(){
        return {
            form: {
                value: '',
            },
            editor: null,
        }
    },
    mounted() {
        this.form.value = this.yaml;
        setTimeout(()=>{
            this.init();
        },300)
    },
    watch: {
        yaml(v){
            this.form.value = v;
            if(this.editor){
                let txt = this.editor.state.doc.toString();
                this.editor.dispatch({
                    changes: {from: 0, to:txt.length, insert:this.form.value}
                });
            }
        }
    },
    methods: {
        getValue(){
            this.form.value = this.editor.state.doc.toString();
            return this.form.value;
        },
        submit(){
            this.$emit('submit', this.getValue());
        },
        cancel(){
            this.$emit('cancel');
        },
        write(value){
            this.form.value = value;
            let txt = this.editor.state.doc.toString();
            this.editor.dispatch({
                changes: {from:0, to:txt.length, insert:value}
            });
        },
        init(){
            if(this.detail){return}
            let height = document.getElementById('yamleditordom').offsetHeight - (this.nofooter===true? 0 : 54);
            
            let myTheme = EditorView.theme({
                "&": {
                    // color: ivory,
                    // backgroundColor: background,
                    height: height + "px"
                },
            }, {dark: false});
            
            if(!this.editor){
                document.getElementById("editor_textarea").innerHTML = "";
                this.editor = new EditorView({
                    doc: "",
                    extensions: [
                        basicSetup,
                        myTheme,
                        yaml(),
                        // keymap.of([indentWithTab]),
                        EditorView.editable.of(!this.disabled),
                        this.createTabKeymap(),
                    ],
                    parent: document.getElementById("editor_textarea"),
                });

                var observe=new MutationObserver(()=>{
                    let txt = this.editor.state.doc.toString();
                    if(txt==this.form.value){return}
                    this.form.value = txt;
                });
                observe.observe(document.getElementById("editor_textarea"),{childList:true, characterData:true, subtree:true});

            }

            let txt = this.editor.state.doc.toString();
            if(!txt && this.form.value){
                this.editor.dispatch({
                    changes: {from: 0, insert:this.form.value}
                });
                let txt = this.editor.state.doc.toString();
                this.form.value = txt;
            }
        },
        // 定义Tab键扩展
        createTabKeymap(){
            const INDENT_SPACES = '  ';
            const INDENT_LEN = 2;
            return keymap.of([
                // 1. Tab键：单行插入缩进+光标后移 | 多行缩进+光标移到最后一行末尾
                {
                key: 'Tab',
                run: (view) => {
                    const { state, dispatch } = view;
                    const { selection } = state;

                    // 场景1：无选中（单行）
                    if (selection.main.empty) {
                    const currentPos = selection.main.from;
                    dispatch({
                        // 插入缩进内容
                        changes: { from: currentPos, to: currentPos, insert: INDENT_SPACES },
                        // 光标后移缩进长度
                        selection: {
                        anchor: currentPos + INDENT_LEN,
                        head: currentPos + INDENT_LEN
                        }
                    });
                    return true;
                    }

                    // 场景2：有选中（多行）
                    const ranges = selection.ranges;
                    const changes = [];
                    let lastLineEnd = 0; // 记录最后一行的结束位置

                    ranges.forEach(range => {
                    const startLine = state.doc.lineAt(range.from);
                    const endLine = state.doc.lineAt(range.to);
                    lastLineEnd = endLine.to; // 更新最后一行位置

                    // 遍历选中的每一行，在行首插入缩进
                    for (let lineNum = startLine.number; lineNum <= endLine.number; lineNum++) {
                        const line = state.doc.line(lineNum);
                        changes.push({ from: line.from, insert: INDENT_SPACES });
                    }
                    });

                    dispatch({
                    changes,
                    // 光标移到最后一行缩进后的位置
                    selection: {
                        anchor: lastLineEnd + INDENT_LEN,
                        head: lastLineEnd + INDENT_LEN
                    }
                    });
                    return true;
                },
                preventDefault: true // 阻止浏览器默认Tab行为（焦点切换）
                },

                // 2. Shift+Tab键：单行取消缩进+光标前移 | 多行取消缩进+光标移到最后一行末尾
                {
                key: 'Shift-Tab',
                run: (view) => {
                    const { state, dispatch } = view;
                    const { selection } = state;

                    // 场景1：无选中（单行）
                    if (selection.main.empty) {
                    const currentPos = selection.main.from;
                    const line = state.doc.lineAt(currentPos);
                    const lineStart = line.from;
                    const lineText = line.text.slice(0, currentPos - lineStart);

                    // 检查行首是否有缩进可取消
                    if (lineText.startsWith(INDENT_SPACES)) {
                        dispatch({
                        // 删除行首的缩进
                        changes: { from: lineStart, to: lineStart + INDENT_LEN },
                        // 光标前移缩进长度（若光标在缩进后，否则移到行首）
                        selection: {
                            anchor: Math.max(currentPos - INDENT_LEN, lineStart),
                            head: Math.max(currentPos - INDENT_LEN, lineStart)
                        }
                        });
                        return true;
                    }
                    return false; // 无缩进可取消，不阻止默认行为
                    }

                    // 场景2：有选中（多行）
                    const ranges = selection.ranges;
                    const changes = [];
                    let lastLineEnd = 0;

                    ranges.forEach(range => {
                    const startLine = state.doc.lineAt(range.from);
                    const endLine = state.doc.lineAt(range.to);
                    lastLineEnd = endLine.to;

                    // 遍历选中的每一行，删除行首的缩进（若有）
                    for (let lineNum = startLine.number; lineNum <= endLine.number; lineNum++) {
                        const line = state.doc.line(lineNum);
                        // 仅当行首有缩进时删除
                        if (line.text.startsWith(INDENT_SPACES)) {
                        changes.push({ from: line.from, to: line.from + INDENT_LEN });
                        }
                    }
                    });

                    // 有缩进可删除时才执行dispatch
                    if (changes.length > 0) {
                    dispatch({
                        changes,
                        // 光标移到最后一行取消缩进后的位置
                        selection: {
                        anchor: Math.max(lastLineEnd - INDENT_LEN, state.doc.lineAt(lastLineEnd).from),
                        head: Math.max(lastLineEnd - INDENT_LEN, state.doc.lineAt(lastLineEnd).from)
                        }
                    });
                    return true;
                    }
                    return false;
                },
                preventDefault: true
                }
            ]);
        }
    }
}
</script>

<style scoped>
pre{margin:0; margin:0; border:0; outline:0; display:block; border-radius:0 4px 4px 0;}
pre code{margin:0; border:0; padding:14px; outline:0; line-height:26px; border-radius:0 4px 4px 0;}
.copybtn{position:absolute; right:20px; top:12px;}
</style>