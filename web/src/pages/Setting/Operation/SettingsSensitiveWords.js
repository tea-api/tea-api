import React, { useEffect, useState, useRef } from 'react';
import { Button, Col, Form, Row, Spin, Tag } from '@douyinfe/semi-ui';
import {
  compareObjects,
  API,
  showError,
  showSuccess,
  showWarning,
} from '../../../helpers';
import { useTranslation } from 'react-i18next';

export default function SettingsSensitiveWords(props) {
  const { t } = useTranslation();
  const [loading, setLoading] = useState(false);
  const [inputs, setInputs] = useState({
    CheckSensitiveEnabled: false,
    CheckSensitiveOnPromptEnabled: false,
    SensitiveWords: '',
  });
  const refForm = useRef();
  const [inputsRow, setInputsRow] = useState(inputs);

  function onSubmit() {
    const updateArray = compareObjects(inputs, inputsRow);
    if (!updateArray.length) return showWarning(t('你似乎并没有修改什么'));
    const requestQueue = updateArray.map((item) => {
      let value = '';
      if (typeof inputs[item.key] === 'boolean') {
        value = String(inputs[item.key]);
      } else {
        value = inputs[item.key];
      }
      return API.put('/api/option/', {
        key: item.key,
        value,
      });
    });
    setLoading(true);
    Promise.all(requestQueue)
      .then((res) => {
        if (requestQueue.length === 1) {
          if (res.includes(undefined)) return;
        } else if (requestQueue.length > 1) {
          if (res.includes(undefined))
            return showError(t('部分保存失败，请重试'));
        }
        showSuccess(t('保存成功'));
        props.refresh();
      })
      .catch(() => {
        showError(t('保存失败，请重试'));
      })
      .finally(() => {
        setLoading(false);
      });
  }

  useEffect(() => {
    // 确保 props.options 存在且为对象
    if (!props.options || typeof props.options !== 'object') {
      console.warn('props.options is invalid:', props.options);
      return;
    }

    const currentInputs = {};
    for (let key in props.options) {
      if (Object.keys(inputs).includes(key)) {
        currentInputs[key] = props.options[key];
      }
    }

    // 确保 currentInputs 不为空，合并默认值
    setInputs(prevInputs => ({ ...prevInputs, ...currentInputs }));
    setInputsRow(structuredClone({ ...inputs, ...currentInputs }));
    if (refForm.current) refForm.current.setValues({ ...inputs, ...currentInputs });
  }, [props.options]);
  return (
    <>
      <Spin spinning={loading}>
        <Form
          values={inputs}
          getFormApi={(formAPI) => (refForm.current = formAPI)}
          style={{ marginBottom: 15 }}
        >
          <Form.Section text={t('屏蔽词过滤设置')}>
            <Row gutter={16}>
              <Col xs={24} sm={12} md={8} lg={8} xl={8}>
                <Form.Switch
                  field={'CheckSensitiveEnabled'}
                  label={t('启用屏蔽词过滤功能')}
                  size='default'
                  checkedText='｜'
                  uncheckedText='〇'
                  onChange={(value) => {
                    setInputs((prevInputs) => ({
                      ...(prevInputs || {}),
                      CheckSensitiveEnabled: value,
                    }));
                  }}
                />
              </Col>
              <Col xs={24} sm={12} md={8} lg={8} xl={8}>
                <Form.Switch
                  field={'CheckSensitiveOnPromptEnabled'}
                  label={t('启用 Prompt 检查')}
                  size='default'
                  checkedText='｜'
                  uncheckedText='〇'
                  onChange={(value) =>
                    setInputs((prevInputs) => ({
                      ...(prevInputs || {}),
                      CheckSensitiveOnPromptEnabled: value,
                    }))
                  }
                />
              </Col>
            </Row>
            <Row>
              <Col xs={24} sm={12} md={8} lg={8} xl={8}>
                <Form.TextArea
                  label={t('屏蔽词列表')}
                  extraText={t('一行一个屏蔽词，不需要符号分割')}
                  placeholder={t('一行一个屏蔽词，不需要符号分割')}
                  field={'SensitiveWords'}
                  onChange={(value) =>
                    setInputs((prevInputs) => ({
                      ...(prevInputs || {}),
                      SensitiveWords: value,
                    }))
                  }
                  style={{ fontFamily: 'JetBrains Mono, Consolas' }}
                  autosize={{ minRows: 6, maxRows: 12 }}
                />
              </Col>
            </Row>
            <Row>
              <Button size='default' onClick={onSubmit}>
                {t('保存屏蔽词过滤设置')}
              </Button>
            </Row>
          </Form.Section>
        </Form>
      </Spin>
    </>
  );
}
