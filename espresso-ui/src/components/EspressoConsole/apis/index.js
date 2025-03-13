import axios from 'axios';

const BASE_URL = 'http://localhost:8081'; // Adjust this based on your API base URL
const STATUS_FAILED = 'failed'; // Assuming this constant exists elsewhere in your code

export const getTemplateListing = async () => {
    try {
        const axiosResponse = await axios.get(`${BASE_URL}/list-templates`);
        const res = axiosResponse.data;
        if (res?.status === STATUS_FAILED) {
            throw new Error(res.message || 'Failed to get template list');
        }
        return res;
    } catch (e) {
        if (axios.isAxiosError(e)) {
            throw new Error(e.response?.data?.status?.message || e.message || 'API request failed');
        }
        throw new Error(e?.message || 'Failed to get template list');
    }
};

export const getTemplateHtmlAndJson = async id => {
    try {
        const axiosResponse = await axios.get(`${BASE_URL}/get-template`, {
            params: { template_id: id }
        });
        const res = axiosResponse.data;
        if (res?.status === STATUS_FAILED) {
            throw new Error(res.message || 'Failed to get template HTML and JSON');
        }
        return res;
    } catch (e) {
        if (axios.isAxiosError(e)) {
            throw new Error(e.response?.data?.status?.message || e.message || 'API request failed');
        }
        throw new Error(e?.message || 'Failed to get template HTML and JSON');
    }
};

export const createPdfTemplate = async requestBody => {
    try {
        const axiosResponse = await axios.post(`${BASE_URL}/create-template`, requestBody, {
            headers: {
                'Content-Type': 'application/json',
            }
        });
        
        const res = axiosResponse.data;
        if (res?.status === STATUS_FAILED) {
            throw new Error(res.message || 'PDF template creation failed');
        }

        const resp = {
            status: {
                status: 'success',
                message: 'Template created successfully',
            }
        }
        return resp;
    } catch (e) {
        if (axios.isAxiosError(e)) {
            throw new Error(e.response?.data?.status?.message || e.message || 'API request failed');
        }
        throw new Error(e?.message || 'Failed to create PDF template');
    }
};

export const generatePdfReq = async requestBody => {
    try {
        const response = await axios.post(`${BASE_URL}/generate-pdf-stream`, requestBody, {
            responseType: 'arraybuffer',
            headers: {
                'Content-Type': 'application/json',
                'Accept': 'application/pdf'
            }
        });
        
        return response.data;
    } catch (e) {
        if (axios.isAxiosError(e)) {
            if (e.response?.data instanceof ArrayBuffer) {
                const decoder = new TextDecoder('utf-8');
                const errorText = decoder.decode(e.response.data);
                try {
                    const errorObj = JSON.parse(errorText);
                    throw new Error(errorObj.message || 'PDF generation failed');
                } catch (jsonErr) {
                    throw new Error('Failed to generate PDF', jsonErr);
                }
            } else {
                throw new Error(e.response?.data?.status?.message || e.message || 'API request failed');
            }
        }
        throw new Error(e?.message || 'Failed to generate PDF');
    }
};


// check how to remove 
export const getTemplateById = async id => {
    try {
        const axiosResponse = await axios.get(`${BASE_URL}/getTemplateById`, {
            params: { template_id: id }
        });
        const res = axiosResponse.data;
        if (res?.status === STATUS_FAILED) {
            throw new Error(res.message || 'Failed to get template');
        }
        return res;
    } catch (e) {
        if (axios.isAxiosError(e)) {
            throw new Error(e.response?.data?.message || e.message || 'API request failed');
        }
        throw new Error(e?.message || 'Failed to get template');
    }
};


export const updatePdfTemplate = async requestBody => {
    try {
        const axiosResponse = await axios.put(`${BASE_URL}/updatePdfTemplate`, requestBody, {
            headers: {
                'Content-Type': 'application/json',
            }
        });
        const res = axiosResponse.data;
        if (res?.status === STATUS_FAILED) {
            throw new Error(res.message || 'Failed to update PDF template');
        }
        return res;
    } catch (e) {
        if (axios.isAxiosError(e)) {
            throw new Error(e.response?.data?.message || e.message || 'API request failed');
        }
        throw new Error(e?.message || 'Failed to update PDF template');
    }
};

export const updateTemplateStatus = async requestBody => {
    try {
        const axiosResponse = await axios.put(`${BASE_URL}/updateTemplateStatus`, requestBody, {
            headers: {
                'Content-Type': 'application/json',
            }
        });
        const res = axiosResponse.data;
        if (res?.status === STATUS_FAILED) {
            throw new Error(res.message || 'Failed to update template status');
        }
        return res;
    } catch (e) {
        if (axios.isAxiosError(e)) {
            throw new Error(e.response?.data?.message || e.message || 'API request failed');
        }
        throw new Error(e?.message || 'Failed to update template status');
    }
};

